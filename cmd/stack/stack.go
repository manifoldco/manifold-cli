package stack

// StackYaml defines the struct definition for the decoded stack.yml
type StackYaml struct {
	// Project is a project label the defines the project that this stack belongs to
	Project string `yaml:"project"`

	// Resources are the Manifold-defined resources in the stack
	Resources map[string]StackResource `yaml:"resources,flow"`
}

// StackResource defines a single resource definition for the a stack, and defines resource
// properties for the stack
type StackResource struct {
	// Title is the descriptive display name for the resource
	Title string `yaml:"title"`

	// Product is the label of the Product for this element in the stack
	Product string `yaml:"product"`

	// Plan is the label of the Plan the resource should use
	Plan string `yaml:"plan"`

	// Region is the label of the Region the resource should use
	Region string `yaml:"region"`
}
