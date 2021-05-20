package util

import (
	"regexp"
	"strings"

	"github.com/docker/docker/api/types/swarm"
)

type constraintType string

const (
	constraintTypeNodeID       constraintType = "node.id"
	constraintTypeNodeHostname constraintType = "node.hostname"
	constraintTypeNodeRole     constraintType = "node.role"
	constraintTypeNodeLabel    constraintType = "node.labels"
)

type operationType string

const (
	operationTypeEqaul    operationType = "=="
	operationTypeNotEqaul operationType = "!="
)

type constraint struct {
	Ct      constraintType
	Op      operationType
	LabelOp operationType
	Key     string
	Value   string
}

func isNodeIdMatch(node swarm.Node, val string) bool {
	return node.ID == val
}

func isNodeHostnameMatch(node swarm.Node, val string) bool {
	return node.Description.Hostname == val
}

func isNodeRoleMatch(node swarm.Node, val string) bool {
	return node.Spec.Role == swarm.NodeRole(val)
}

func isNodeLabelMatch(node swarm.Node, lop operationType, key string, val string) bool {
	return withOperator(lop, (node.Spec.Labels[key] == val))
}

func withOperator(op operationType, b bool) bool {
	if op == operationTypeEqaul {
		return b
	} else {
		return !b
	}
}

func ExtractConstraint(service swarm.Service) (c []constraint) {
	if len(service.Spec.TaskTemplate.Placement.Constraints) == 0 {
		return c
	}
	for _, sc := range service.Spec.TaskTemplate.Placement.Constraints {
		var nc constraint

		if strings.HasPrefix(sc, string(constraintTypeNodeID)) {
			nc.Ct = constraintTypeNodeID
		} else if strings.HasPrefix(sc, string(constraintTypeNodeHostname)) {
			nc.Ct = constraintTypeNodeHostname
		} else if strings.HasPrefix(sc, string(constraintTypeNodeRole)) {
			nc.Ct = constraintTypeNodeRole
		} else if strings.HasPrefix(sc, string(constraintTypeNodeLabel)) {
			nc.Ct = constraintTypeNodeLabel
		} else {
			continue
		}

		re := regexp.MustCompile(`([a-z\.]+) *([=!]+) *(.+)`)
		split := re.FindStringSubmatch(sc)
		if split[2] == string(operationTypeEqaul) {
			nc.Op = operationTypeEqaul
		} else {
			nc.Op = operationTypeNotEqaul
		}

		if nc.Ct == constraintTypeNodeLabel {
			lre := regexp.MustCompile(`([a-z\.]+) *([=!]+)`)
			k := strings.TrimPrefix(split[0], string(constraintTypeNodeLabel)+".")
			labelSplit := lre.FindStringSubmatch(k)
			nc.LabelOp = operationType(labelSplit[2])
			nc.Key = labelSplit[1]
		}

		nc.Value = split[3]

		c = append(c, nc)
	}
	return c
}

type ExpectedNodeFilter func(swarm.Service) []swarm.Node

func ConstraintFilter(nodes []swarm.Node) ExpectedNodeFilter {
	return func(service swarm.Service) []swarm.Node {
		cs := ExtractConstraint(service)
		if len(cs) == 0 {
			return nodes
		}
		var nl []swarm.Node
		for _, node := range nodes {
			nodeMatch := true
			for _, c := range cs {
				if c.Ct == constraintTypeNodeID {
					if !withOperator(c.Op, isNodeIdMatch(node, c.Value)) {
						nodeMatch = false
					}
				} else if c.Ct == constraintTypeNodeHostname {
					if !withOperator(c.Op, isNodeHostnameMatch(node, c.Value)) {
						nodeMatch = false
					}
				} else if c.Ct == constraintTypeNodeRole {
					if !withOperator(c.Op, isNodeRoleMatch(node, c.Value)) {
						nodeMatch = false
					}
				} else if c.Ct == constraintTypeNodeLabel {
					if !withOperator(c.Op, isNodeLabelMatch(node, c.LabelOp, c.Key, c.Value)) {
						nodeMatch = false
					}
				}
			}
			if nodeMatch {
				nl = append(nl, node)
				continue
			}
		}
		return nl
	}
}
