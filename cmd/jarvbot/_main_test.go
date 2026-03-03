package main

import (
	"reflect"
	"regexp"
	"testing"
	"time"
)

func Test_processTimedCommand(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name  string
		args  args
		want  time.Duration
		want1 string
	}{
		{"Nothing", args{""}, time.Duration(0), ""},
		{"One hour", args{"!remindme 1h pls"}, time.Hour, "pls"},
		{"One minute", args{"!remindme 1m msg"}, time.Minute, "msg"},
		{"One second", args{"!remindme 1s msg"}, time.Second, "msg"},
		{"Five hours and twenty minutes", args{"!shutdown 5h 20m msg pls"}, time.Hour*5 + time.Minute*20, "msg pls"},
		{"Six hours and six seconds", args{"!shutdown 6h 6s"}, time.Hour*6 + time.Second*6, ""},
		{"Three hours fifty mins two secs", args{"!a   3h 50m  2s   blabla"}, time.Hour*3 + time.Minute*50 + time.Second*2, "blabla"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := processTimedCommand(tt.args.s)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processRemindmeCommand() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("processRemindmeCommand() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_extractTimeUnit(t *testing.T) {
	type args struct {
		s  string
		re *regexp.Regexp
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 string
	}{
		{"One hour", args{"1h blabla", stringHoursRegex}, 1, "blabla"},
		{"Two minutes hour", args{"2m message", stringMinsRegex}, 2, "message"},
		{"Fifty seconds", args{"50s", stringSecsRegex}, 50, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := extractTimeUnit(tt.args.s, tt.args.re)
			if got != tt.want {
				t.Errorf("extractTimeUnit() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("extractTimeUnit() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
