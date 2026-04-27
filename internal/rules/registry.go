package rules

type Registry struct {
	rules []Rule
}

func NewRegistry() *Registry {
	return &Registry{
		rules: make([]Rule, 0),
	}
}

func (r *Registry) Register(rule Rule) {
	r.rules = append(r.rules, rule)
}

func (r *Registry) GetAll() []Rule {
	return r.rules
}

func (r *Registry) GetByName(name string) Rule {
	for _, rule := range r.rules {
		if rule.Name() == name {
			return rule
		}
	}
	return nil
}
