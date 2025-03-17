package service

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/gofiber/fiber/v2/log"
	"liteide-backend/ent/image"
	"liteide-backend/ent/property"
	"liteide-backend/svc"
	"path"
	"strconv"
	"time"
)

// CreateContainer 创建一个新的 Docker 容器
// - `ctx`：请求的上下文
// - `userId`：创建容器的用户 ID
// - `workspaceId`：关联的工作区 ID
// - 返回容器 ID 和错误信息（如果有）
func CreateContainer(ctx context.Context, userId int, workspaceId int) (*int, error) {
	// 获取工作区信息
	workspaceInstance, err := svc.SVC.Database.Workspace.Get(ctx, workspaceId)
	if err != nil {
		return nil, err
	}

	// 查询该工作区对应的镜像信息
	imageInstance, err := svc.SVC.Database.Image.Query().
		Where(image.Language(workspaceInstance.Language)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	// 在数据库中创建容器记录（状态：Pending）
	container, err := svc.SVC.Database.Container.Create().
		SetUserID(userId).
		SetImage(imageInstance).
		SetWorkspace(workspaceInstance).
		SetContainerStatus(property.ContainerStatusPending).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// 设置 Swarm 任务副本数
	replicas := uint64(1)

	// 定义 Swarm 服务配置（相当于 Docker Service）
	serviceSpec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: svc.SVC.AppConfig.ContainerServicePrefix + strconv.Itoa(container.ID),
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image: imageInstance.ImageName,
				TTY:   true,
				Dir:   "/workspace",
				Mounts: []mount.Mount{
					{
						Type: mount.TypeBind, // 绑定本地目录到容器
						Source: path.Join(
							svc.SVC.AppConfig.DataDirectory,
							"workspace",
							workspaceInstance.UUID.String(),
						),
						Target: "/workspace",
					},
				},
			},
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &replicas,
			},
		},
	}

	// 创建 Swarm 服务（相当于 Docker 容器）
	service, err := svc.SVC.Docker.ServiceCreate(ctx, serviceSpec, types.ServiceCreateOptions{})
	if err != nil {
		// 如果创建失败，则更新数据库状态为 "Removed"
		if err := svc.SVC.Database.Container.UpdateOne(container).
			SetContainerStatus(property.ContainerStatusRemoved).
			Exec(ctx); err != nil {
			log.Errorf("failed to update container status: %v", err)
		}
		return nil, err
	}

	// TODO: 监听容器启动状态
	err = svc.SVC.Database.Container.UpdateOne(container).
		SetContainerStatus(property.ContainerStatusUp).
		SetContainerID(service.ID).
		Exec(ctx)

	if err != nil {
		// 如果数据库更新失败，删除创建的 Swarm 服务
		if err := svc.SVC.Docker.ServiceRemove(ctx, service.ID); err != nil {
			log.Errorf("failed to remove service: %v", err)
			_ = svc.SVC.Database.Container.UpdateOne(container).
				SetContainerStatus(property.ContainerStatusError).
				Exec(ctx)
		}
		return nil, err
	}

	log.Debugf("service created: %v", service.ID)
	return &container.ID, nil
}

// RemoveContainer 删除 Docker 容器
// - `ctx`：请求的上下文
// - `containerId`：要删除的容器 ID
// - 返回错误信息（如果有）
func RemoveContainer(ctx context.Context, containerId int) error {
	// 获取容器信息
	container, err := svc.SVC.Database.Container.Get(ctx, containerId)
	if err != nil {
		return err
	}

	// 只有状态为 "Up" 且存在 `ContainerID` 的容器才能删除
	if container.ContainerStatus != property.ContainerStatusUp || container.ContainerID == nil {
		return fmt.Errorf("container is not running")
	}

	// 调用 Docker API 删除 Swarm 服务
	err = svc.SVC.Docker.ServiceRemove(ctx, *container.ContainerID)
	if err != nil {
		return err
	}

	// 更新数据库状态为 "Removed" 并清除 `ContainerID`
	return svc.SVC.Database.Container.UpdateOne(container).
		SetContainerStatus(property.ContainerStatusRemoved).
		ClearContainerID().
		SetExitTime(time.Now()). // 记录删除时间
		Exec(ctx)
}

// AttachContainer 附加到正在运行的 Docker 容器
// - `ctx`：请求的上下文
// - `containerId`：要附加的容器 ID
// - 返回 WebSocket 连接（HijackedResponse）和错误信息（如果有）
func AttachContainer(ctx context.Context, containerId int) (*types.HijackedResponse, error) {
	// 获取容器信息
	container, err := svc.SVC.Database.Container.Get(ctx, containerId)
	if err != nil {
		return nil, err
	}

	// 只有状态为 "Up" 且存在 `ContainerID` 的容器才能附加
	if container.ContainerStatus != property.ContainerStatusUp || container.ContainerID == nil {
		return nil, fmt.Errorf("container is not running")
	}

	// 在 Docker Swarm 中查找容器实例
	instanceList, err := svc.SVC.Docker.ContainerList(ctx, types.ContainerListOptions{
		Filters: func() filters.Args {
			filterArgs := filters.NewArgs()
			filterName := svc.SVC.AppConfig.ContainerServicePrefix + strconv.Itoa(container.ID)
			filterArgs.Add("name", filterName)
			return filterArgs
		}(),
	})
	if err != nil {
		return nil, err
	}
	if len(instanceList) == 0 {
		return nil, fmt.Errorf("no container found")
	}

	// 获取找到的第一个容器实例
	instance := instanceList[0]
	log.Debugf("instance: %v", instance)

	// 创建 Docker Exec 进程（进入容器 /bin/sh）
	execConfig, err := svc.SVC.Docker.ContainerExecCreate(ctx, instance.ID, types.ExecConfig{
		AttachStdin:  true,                // 允许输入
		AttachStdout: true,                // 允许输出
		AttachStderr: true,                // 允许错误输出
		Cmd:          []string{"/bin/sh"}, // 运行 `/bin/sh`
		Tty:          true,                // 启用 TTY 模式
	})
	if err != nil {
		return nil, err
	}

	// 附加到 Exec 进程，建立 WebSocket 连接
	conn, err := svc.SVC.Docker.ContainerExecAttach(ctx, execConfig.ID, types.ExecStartCheck{Detach: false, Tty: true})
	if err != nil {
		return nil, err
	}
	return &conn, nil
}
