module alice

go 1.13

replace logging => ../../../logging

require (
	github.com/leesper/go_rng v0.0.0-20190531154944-a612b043e353
	logging v0.0.0-00010101000000-000000000000
)
