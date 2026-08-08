package awssync

import (
	"context"
	"fmt"
	"sort"

	"sshh/internal/model"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// rawInstance is an EC2 instance's fields relevant to inventory sync, before
// name disambiguation.
type rawInstance struct {
	name string
	host string
	id   string
}

// FetchAccessibleInstances returns EC2 instances from the given AWS CLI
// profile that are running and tagged sshh_accessible=True, using whatever
// credentials the profile resolves to (shared credentials file, SSO, etc.).
// Instances without a Name tag or a private IP are skipped. Instances that
// share a Name tag (common in ECS/ASG-managed clusters, where every member
// gets the same tag) are disambiguated by appending their instance ID so
// none of them are dropped.
func FetchAccessibleInstances(ctx context.Context, profile string) ([]model.Server, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config for profile %q: %w", profile, err)
	}

	client := ec2.NewFromConfig(cfg)
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{Name: strPtr("instance-state-name"), Values: []string{"running"}},
			{Name: strPtr("tag:sshh_accessible"), Values: []string{"True"}},
		},
	}

	var raw []rawInstance

	paginator := ec2.NewDescribeInstancesPaginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describing EC2 instances for profile %q: %w", profile, err)
		}
		for _, res := range page.Reservations {
			for _, inst := range res.Instances {
				name := instanceNameTag(inst)
				if name == "" {
					continue
				}
				if inst.PrivateIpAddress == nil || *inst.PrivateIpAddress == "" {
					continue
				}
				id := ""
				if inst.InstanceId != nil {
					id = *inst.InstanceId
				}
				raw = append(raw, rawInstance{name: name, host: *inst.PrivateIpAddress, id: id})
			}
		}
	}

	nameCount := make(map[string]int, len(raw))
	for _, r := range raw {
		nameCount[r.name]++
	}

	servers := make([]model.Server, 0, len(raw))
	for _, r := range raw {
		name := r.name
		if nameCount[name] > 1 && r.id != "" {
			name = name + "-" + r.id
		}
		servers = append(servers, model.Server{
			Name: name,
			Host: r.host,
			Port: 22,
		})
	}

	sort.Slice(servers, func(i, j int) bool { return servers[i].Name < servers[j].Name })
	return servers, nil
}

// instanceNameTag returns the value of the instance's "Name" tag, or "" if absent.
func instanceNameTag(inst types.Instance) string {
	for _, t := range inst.Tags {
		if t.Key != nil && *t.Key == "Name" && t.Value != nil {
			return *t.Value
		}
	}
	return ""
}

func strPtr(s string) *string { return &s }
