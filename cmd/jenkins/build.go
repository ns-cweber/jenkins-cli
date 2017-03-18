package main

import jenkins "github.com/ns-cweber/jenkins-cli"

func get(b jenkins.Build, s string) string {
	switch s {
	case "number":
		return b.Number
	case "worker":
		return b.BuiltOn
	case "status":
		return string(b.Result)
	default:
		for _, action := range b.Actions {
			if action.Class == jenkins.ActionClassParameters {
				return action.Parameters.Get(s)
			}
		}
		return ""
	}
}
