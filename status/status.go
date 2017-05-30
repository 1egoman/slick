package status

import (
	"fmt"
)

type StatusType int

const (
	STATUS_ERROR StatusType = iota
	STATUS_LOG
	STATUS_INFO
)

type Status struct {
	Message string
	Type    StatusType
	Show    bool
}

func (s *Status) Printf(format string, args ...interface{}) {
	s.Message = fmt.Sprintf(format, args...)
	s.Type = STATUS_LOG
	s.Show = true
}

func (s *Status) Errorf(format string, args ...interface{}) {
	s.Message = fmt.Sprintf(format, args...)
	s.Type = STATUS_ERROR
	s.Show = true
}

func (s *Status) Clear() {
	s.Show = false
}
