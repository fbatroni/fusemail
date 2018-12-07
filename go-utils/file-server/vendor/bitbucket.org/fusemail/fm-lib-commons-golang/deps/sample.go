package deps

// Provides Sample struct to check health as DEMO only.

// Sample provides health struct.
type Sample struct {
	Any string `json:"any"`

	// Sample IGNORE during json marshal,
	// otherwise panics due to channel usage.
	Private chan bool `json:"-"`
}

// Check checks health.
// Returns map (optional config/state) and error (nil if healthy).
func (d *Sample) Check() (map[string]interface{}, error) {
	var state map[string]interface{} // nil = no required state.
	var err error                    // nil = healthy.
	return state, err
}
