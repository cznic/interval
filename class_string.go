// generated by stringer -type Class; DO NOT EDIT

package interval

import "fmt"

const _Class_name = "UnboundedEmptyDegenerateOpenClosedLeftOpenLeftClosedLeftBoundedOpenLeftBoundedClosedRightBoundedOpenRightBoundedClosednClasses"

var _Class_index = [...]uint8{0, 9, 14, 24, 28, 34, 42, 52, 67, 84, 100, 118, 126}

func (i Class) String() string {
	if i < 0 || i >= Class(len(_Class_index)-1) {
		return fmt.Sprintf("Class(%d)", i)
	}
	return _Class_name[_Class_index[i]:_Class_index[i+1]]
}
