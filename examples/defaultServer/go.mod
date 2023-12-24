module defaultServer

go 1.20

require tcp v0.0.0

replace (
	tcp  => ../../../tcp
)
