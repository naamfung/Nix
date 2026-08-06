//go:build darwin

package main

/*
#include <stdint.h>
#include <dispatch/dispatch.h>

extern void inxDesktopMainHeartbeat(void);

static dispatch_source_t inx_main_heartbeat_timer;

static void inx_main_heartbeat_handler(void *ctx) {
	inxDesktopMainHeartbeat();
}

static void inx_start_main_heartbeat(uint64_t interval_ms) {
	if (inx_main_heartbeat_timer != NULL) {
		return;
	}
	inx_main_heartbeat_timer = dispatch_source_create(DISPATCH_SOURCE_TYPE_TIMER, 0, 0, dispatch_get_main_queue());
	dispatch_set_context(inx_main_heartbeat_timer, NULL);
	dispatch_source_set_event_handler_f(inx_main_heartbeat_timer, inx_main_heartbeat_handler);
	dispatch_source_set_timer(inx_main_heartbeat_timer, dispatch_time(DISPATCH_TIME_NOW, 0), interval_ms * NSEC_PER_MSEC, 100 * NSEC_PER_MSEC);
	dispatch_resume(inx_main_heartbeat_timer);
}

static void inx_stop_main_heartbeat(void) {
	if (inx_main_heartbeat_timer == NULL) {
		return;
	}
	dispatch_source_cancel(inx_main_heartbeat_timer);
	inx_main_heartbeat_timer = NULL;
}
*/
import "C"

import "time"

func mainThreadWatchdogSupported() bool {
	return true
}

func startNativeMainThreadHeartbeat(intervalMS uint64) {
	C.inx_start_main_heartbeat(C.uint64_t(intervalMS))
}

func stopNativeMainThreadHeartbeat() {
	C.inx_stop_main_heartbeat()
}

//export inxDesktopMainHeartbeat
func inxDesktopMainHeartbeat() {
	recordMainThreadHeartbeat(time.Now())
}
