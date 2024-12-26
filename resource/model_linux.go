package resource

var Exclusive = MResource{
	"in":        {Local: "in", Remote: "in_linux_amd64", RemoteArm: "in_linux_arm"},
	"upgrade":   {Local: "in_upgrade", Remote: "upgrade_linux_amd64", RemoteArm: "upgrade_linux_arm", Key: []string{"in_upgrade"}},
	"forward":   {Local: "forward", Remote: "forward_linux_amd64", RemoteArm: "forward_linux_arm"},
	"edge":      {Local: "edge", Remote: "edge_linux_amd64", RemoteArm: "edge_linux_arm"},
	"edge_mini": {Local: "edge_mini", Remote: "edge_mini_linux_amd64", RemoteArm: "edge_mini_linux_arm"},
	"notice":    {Local: "notice", Remote: "notice_linux_amd64", RemoteArm: "notice_linux_arm"},
}
