package segment

// DefaultRegistry creates a Registry pre-populated with the 16 segment groups.
func DefaultRegistry() *Registry {
	r := NewRegistry()

	r.Register(&ModelGroup{})
	r.Register(&GitGroup{})
	r.Register(&ContextGroup{})
	r.Register(&TokensGroup{})
	r.Register(&CostGroup{})
	r.Register(&UsageGroup{})
	r.Register(&WeeklyGroup{})
	r.Register(&VersionGroup{})
	r.Register(&SessionGroup{})
	r.Register(&SpeedGroup{})
	r.Register(&DiffGroup{})
	r.Register(&ActivityGroup{})
	r.Register(&LiveGroup{})
	r.Register(&CWDGroup{})
	r.Register(&EnvGroup{})
	r.Register(&ClockGroup{})

	return r
}
