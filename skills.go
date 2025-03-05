package main

type Ultimates struct {
	Value int
	Move  int
}

func ultimate(className string) *Ultimates {
	var classSkill Ultimates

	switch className {
	case "Некромант":
		classSkill = Ultimates{
			Value: 50,
		}
	}
	return &classSkill
}
