package ui

// The constructors main needs, exported without exposing the view internals.

type View = view
type Target = target

func NewProjectsView(env *Env) View               { return newProjectsView(env) }
func NewProjectView(env *Env, prefix string) View { return newProjectView(env, prefix) }
func NewTaskView(env *Env, t Target) View         { return newTaskView(env, t) }
