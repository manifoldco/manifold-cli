package stack

// StackYaml defines the struct definition for the decoded stack.yaml
type StackYaml struct {
	// Project is a project label the defines the project that this stack belongs to
	Project string `yaml:"project"`

	// Resources are the Manifold-defined resources in the stack
	Resources []StackResource `yaml:"resources,flow"`
}

// StackResource defines a single resource definition for the a stack, and defines resource
// properties for the stack
type StackResource struct {
	// Resource it the label of the resource for this element in the stack
	Resource string `yaml:"resource"`

	// Product is the label of the Product for this element in the stack
	Product string `yaml:"product"`

	// Plan is the label of the Plan the resource should use
	Plan string `yaml:"plan"`
}
