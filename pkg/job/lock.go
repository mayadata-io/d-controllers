/*
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package job

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "mayadata.io/d-operators/types/job"
)

// LockRunner executes a lock task
type LockRunner struct {
	*Fixture

	LockForever bool
	Task        types.Task
}

func (r *LockRunner) delete() (types.TaskStatus, error) {
	var message = fmt.Sprintf(
		"Delete: Lock %s %s: GVK %s",
		r.Task.Apply.State.GetNamespace(),
		r.Task.Apply.State.GetName(),
		r.Task.Apply.State.GroupVersionKind(),
	)
	client, err := r.dynamicClientset.
		GetClientForAPIVersionAndKind(
			r.Task.Apply.State.GetAPIVersion(),
			r.Task.Apply.State.GetKind(),
		)
	if err != nil {
		return types.TaskStatus{}, err
	}
	err = client.
		Namespace(r.Task.Apply.State.GetNamespace()).
		Delete(
			r.Task.Apply.State.GetName(),
			&metav1.DeleteOptions{},
		)
	if err != nil {
		return types.TaskStatus{}, err
	}
	return types.TaskStatus{
		Phase:   types.TaskStatusPassed,
		Message: message,
	}, nil
}

func (r *LockRunner) create() (types.TaskStatus, error) {
	var message = fmt.Sprintf(
		"Create: Lock %s %s: GVK %s",
		r.Task.Apply.State.GetNamespace(),
		r.Task.Apply.State.GetName(),
		r.Task.Apply.State.GroupVersionKind(),
	)
	client, err := r.dynamicClientset.
		GetClientForAPIVersionAndKind(
			r.Task.Apply.State.GetAPIVersion(),
			r.Task.Apply.State.GetKind(),
		)
	if err != nil {
		return types.TaskStatus{}, err
	}
	_, err = client.
		Namespace(r.Task.Apply.State.GetNamespace()).
		Create(
			r.Task.Apply.State,
			metav1.CreateOptions{},
		)
	if err != nil {
		return types.TaskStatus{}, err
	}
	return types.TaskStatus{
		Phase:   types.TaskStatusPassed,
		Message: message,
	}, nil
}

// Lock acquires the lock and returns unlock
func (r *LockRunner) Lock() (
	types.TaskStatus,
	func() (types.TaskStatus, error),
	error,
) {
	lockstatus, err := r.create()
	if err != nil {
		return types.TaskStatus{}, nil, err
	}
	// build the unlock logic
	var unlock func() (types.TaskStatus, error)
	if r.LockForever {
		unlock = func() (types.TaskStatus, error) {
			// it is a noop if this lock is meant
			// to be present forever
			return types.TaskStatus{
				Phase:   types.TaskStatusPassed,
				Message: "Locked forever",
			}, nil
		}
	} else {
		unlock = r.delete
	}
	return lockstatus, unlock, nil
}

// MustUnlock executes unlock logic without considering
// at any criteria
func (r *LockRunner) MustUnlock() (types.TaskStatus, error) {
	return r.delete()
}
