#!/sbin/openrc-run

name=$RC_SVCNAME
supervisor="supervise-daemon"

output_log="/var/log/${RC_SVCNAME}.log"
error_log="/var/log/${RC_SVCNAME}.err"
pidfile="/run/${RC_SVCNAME}.pid"

command="/usr/bin/alpha"

depend() {
	after net
}