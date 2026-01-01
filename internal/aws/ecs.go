package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

type ECSClient struct {
	client *ecs.Client
}

func NewECSClient(ctx context.Context, profile string) (*ECSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &ECSClient{
		client: ecs.NewFromConfig(cfg),
	}, nil
}

type ECSClusterInfo struct {
	ARN                string
	Name               string
	Status             string
	RunningTasks       int32
	PendingTasks       int32
	ActiveServices     int32
}

func (c *ECSClient) ListClusters(ctx context.Context) ([]ECSClusterInfo, error) {
	output, err := c.client.ListClusters(ctx, &ecs.ListClustersInput{})
	if err != nil {
		return nil, err
	}

	if len(output.ClusterArns) == 0 {
		return nil, nil
	}

	describeOutput, err := c.client.DescribeClusters(ctx, &ecs.DescribeClustersInput{
		Clusters: output.ClusterArns,
	})
	if err != nil {
		return nil, err
	}

	var clusters []ECSClusterInfo
	for _, cl := range describeOutput.Clusters {
		clusters = append(clusters, ECSClusterInfo{
			ARN:            aws.ToString(cl.ClusterArn),
			Name:           aws.ToString(cl.ClusterName),
			Status:         aws.ToString(cl.Status),
			RunningTasks:   cl.RunningTasksCount,
			PendingTasks:   cl.PendingTasksCount,
			ActiveServices: cl.ActiveServicesCount,
		})
	}

	return clusters, nil
}

type ServiceInfo struct {
	ARN            string
	Name           string
	Status         string
	DesiredTasks   int32
	RunningTasks   int32
	LaunchType     string
	TaskDefinition string
}

func (c *ECSClient) ListServices(ctx context.Context, cluster string) ([]ServiceInfo, error) {
	var serviceArns []string
	paginator := ecs.NewListServicesPaginator(c.client, &ecs.ListServicesInput{
		Cluster: aws.String(cluster),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		serviceArns = append(serviceArns, page.ServiceArns...)
	}

	if len(serviceArns) == 0 {
		return nil, nil
	}

	var services []ServiceInfo
	// DescribeServices has a limit of 10
	for i := 0; i < len(serviceArns); i += 10 {
		end := i + 10
		if end > len(serviceArns) {
			end = len(serviceArns)
		}

		describeOutput, err := c.client.DescribeServices(ctx, &ecs.DescribeServicesInput{
			Cluster:  aws.String(cluster),
			Services: serviceArns[i:end],
		})
		if err != nil {
			return nil, err
		}

		for _, s := range describeOutput.Services {
			services = append(services, ServiceInfo{
				ARN:            aws.ToString(s.ServiceArn),
				Name:           aws.ToString(s.ServiceName),
				Status:         aws.ToString(s.Status),
				DesiredTasks:   s.DesiredCount,
				RunningTasks:   s.RunningCount,
				LaunchType:     string(s.LaunchType),
				TaskDefinition: aws.ToString(s.TaskDefinition),
			})
		}
	}

	return services, nil
}

type ECSTaskInfo struct {
	ARN               string
	ID                string
	LastStatus        string
	DesiredStatus     string
	TaskDefinition    string
	LaunchType        string
	CPU               string
	Memory            string
	CreatedAt         string
}

func (c *ECSClient) ListTasks(ctx context.Context, cluster string, serviceName *string) ([]ECSTaskInfo, error) {
	input := &ecs.ListTasksInput{
		Cluster: aws.String(cluster),
	}
	if serviceName != nil {
		input.ServiceName = serviceName
	}

	output, err := c.client.ListTasks(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(output.TaskArns) == 0 {
		return nil, nil
	}

	describeOutput, err := c.client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   output.TaskArns,
	})
	if err != nil {
		return nil, err
	}

	var tasks []ECSTaskInfo
	for _, t := range describeOutput.Tasks {
		id := aws.ToString(t.TaskArn)
		if lastSlash := strings.LastIndex(id, "/"); lastSlash != -1 {
			id = id[lastSlash+1:]
		}

		createdAt := ""
		if t.CreatedAt != nil {
			createdAt = t.CreatedAt.Format("2006-01-02 15:04")
		}

		tasks = append(tasks, ECSTaskInfo{
			ARN:            aws.ToString(t.TaskArn),
			ID:             id,
			LastStatus:     aws.ToString(t.LastStatus),
			DesiredStatus:  aws.ToString(t.DesiredStatus),
			TaskDefinition: aws.ToString(t.TaskDefinitionArn),
			LaunchType:     string(t.LaunchType),
			CPU:            aws.ToString(t.Cpu),
			Memory:         aws.ToString(t.Memory),
			CreatedAt:      createdAt,
		})
	}

	return tasks, nil
}

type TaskDefinitionInfo struct {
	ARN      string
	Family   string
	Revision int32
	Status   string
}

func (c *ECSClient) ListAllTaskDefinitions(ctx context.Context) ([]TaskDefinitionInfo, error) {
	var taskDefs []TaskDefinitionInfo
	paginator := ecs.NewListTaskDefinitionsPaginator(c.client, &ecs.ListTaskDefinitionsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, arn := range page.TaskDefinitionArns {
			family := arn
			var revision int32
			if lastColon := strings.LastIndex(arn, ":"); lastColon != -1 {
				family = arn[strings.LastIndex(arn, "/")+1 : lastColon]
				fmt.Sscanf(arn[lastColon+1:], "%d", &revision)
			}

			taskDefs = append(taskDefs, TaskDefinitionInfo{
				ARN:      arn,
				Family:   family,
				Revision: revision,
				Status:   "ACTIVE",
			})
		}
	}

	return taskDefs, nil
}

func (c *ECSClient) ListTaskDefinitionFamilies(ctx context.Context) ([]string, error) {
	var families []string
	paginator := ecs.NewListTaskDefinitionFamiliesPaginator(c.client, &ecs.ListTaskDefinitionFamiliesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		families = append(families, page.Families...)
	}

	return families, nil
}

func (c *ECSClient) ListTaskDefinitionRevisions(ctx context.Context, family string) ([]TaskDefinitionInfo, error) {
	var taskDefs []TaskDefinitionInfo
	paginator := ecs.NewListTaskDefinitionsPaginator(c.client, &ecs.ListTaskDefinitionsInput{
		FamilyPrefix: aws.String(family),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, arn := range page.TaskDefinitionArns {
			var revision int32
			if lastColon := strings.LastIndex(arn, ":"); lastColon != -1 {
				fmt.Sscanf(arn[lastColon+1:], "%d", &revision)
			}

			taskDefs = append(taskDefs, TaskDefinitionInfo{
				ARN:      arn,
				Family:   family,
				Revision: revision,
				Status:   "ACTIVE",
			})
		}
	}

	return taskDefs, nil
}

func (c *ECSClient) GetTaskDefinitionJSON(ctx context.Context, arn string) (string, error) {
	output, err := c.client.DescribeTaskDefinition(ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(arn),
	})
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(output.TaskDefinition, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

type ECSEventInfo struct {
	ID        string
	CreatedAt string
	Message   string
}

func (c *ECSClient) GetServiceEvents(ctx context.Context, cluster, service string) ([]ECSEventInfo, error) {
	output, err := c.client.DescribeServices(ctx, &ecs.DescribeServicesInput{
		Cluster:  aws.String(cluster),
		Services: []string{service},
	})
	if err != nil {
		return nil, err
	}

	if len(output.Services) == 0 {
		return nil, nil
	}

	var events []ECSEventInfo
	for _, e := range output.Services[0].Events {
		createdAt := ""
		if e.CreatedAt != nil {
			createdAt = e.CreatedAt.Format("15:04:05")
		}
		events = append(events, ECSEventInfo{
			ID:        aws.ToString(e.Id),
			CreatedAt: createdAt,
			Message:   aws.ToString(e.Message),
		})
	}

	return events, nil
}

func (c *ECSClient) StopTask(ctx context.Context, cluster, taskArn string) error {
	_, err := c.client.StopTask(ctx, &ecs.StopTaskInput{
		Cluster: aws.String(cluster),
		Task:    aws.String(taskArn),
		Reason:  aws.String("Restarted from AWS TUI"),
	})
	return err
}

func (c *ECSClient) StopService(ctx context.Context, cluster, service string) error {
	_, err := c.client.UpdateService(ctx, &ecs.UpdateServiceInput{
		Cluster:      aws.String(cluster),
		Service:      aws.String(service),
		DesiredCount: aws.Int32(0),
	})
	return err
}

func (c *ECSClient) RestartService(ctx context.Context, cluster, service string) error {
	_, err := c.client.UpdateService(ctx, &ecs.UpdateServiceInput{
		Cluster:            aws.String(cluster),
		Service:            aws.String(service),
		ForceNewDeployment: true,
	})
	return err
}

func (c *ECSClient) GetLogGroupForTaskDefinition(ctx context.Context, taskDefArn string) (string, error) {
	output, err := c.client.DescribeTaskDefinition(ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefArn),
	})
	if err != nil {
		return "", err
	}

	for _, cd := range output.TaskDefinition.ContainerDefinitions {
		if cd.LogConfiguration != nil && cd.LogConfiguration.LogDriver == types.LogDriverAwslogs {
			if group, ok := cd.LogConfiguration.Options["awslogs-group"]; ok {
				return group, nil
			}
		}
	}

	return "", fmt.Errorf("no CloudWatch log group found in task definition")
}
