package utils

import "sync"

type GlobalVariableStore interface {
	ExistsNode(nodeId string) bool
	SetNodeVariables(nodeId string, variables map[string]string)
	GetVariable(nodeId, name string) (string, bool)
}

type globalVariableNodeMemory struct {
	variables map[string]string
}

type GlobalVariablesMemory struct {
	nodes map[string]*globalVariableNodeMemory
	sync.RWMutex
}

func NewGlobalVariablesMemoryStore() GlobalVariableStore {
	return &GlobalVariablesMemory{
		nodes: make(map[string]*globalVariableNodeMemory),
	}
}

func (store *GlobalVariablesMemory) ExistsNode(nodeId string) bool {
	store.RLock()
	defer store.RUnlock()
	_, ok := store.nodes[nodeId]
	return ok
}

func (store *GlobalVariablesMemory) getNodeStore(nodeId string) (*globalVariableNodeMemory, bool) {
	store.RLock()
	defer store.RUnlock()
	node, ok := store.nodes[nodeId]
	return node, ok
}

func (store *GlobalVariablesMemory) SetNodeVariables(nodeId string, variables map[string]string) {
	store.Lock()
	defer store.Unlock()
	store.nodes[nodeId] = &globalVariableNodeMemory{variables}
}

func (store *GlobalVariablesMemory) GetVariable(nodeId, name string) (string, bool) {
	if node, ok := store.getNodeStore(nodeId); ok {
		return node.GetVariable(name)
	}

	return "", false
}

func (node *globalVariableNodeMemory) GetVariable(name string) (string, bool) {
	value, ok := node.variables[name]
	return value, ok
}
