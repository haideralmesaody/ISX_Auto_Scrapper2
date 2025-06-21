package strategies

// evalRule returns true or false for a single rule given a value.
func evalRule(val float64, r Rule) bool {
	switch r.Operator {
	case ">":
		return val > r.Target
	case "<":
		return val < r.Target
	case ">=":
		return val >= r.Target
	case "<=":
		return val <= r.Target
	case "==":
		return val == r.Target
	case "!=":
		return val != r.Target
	default:
		return false
	}
}

// evalChain walks the slice left-to-right applying AND/OR logic.
// The first rule's Link is ignored (usually empty).
func evalChain(val float64, rules []Rule) bool {
	if len(rules) == 0 {
		return false
	}

	result := evalRule(val, rules[0])
	for i := 1; i < len(rules); i++ {
		cond := evalRule(val, rules[i])
		if rules[i].Link == "OR" {
			result = result || cond
		} else { // default AND
			result = result && cond
		}
	}
	return result
}
