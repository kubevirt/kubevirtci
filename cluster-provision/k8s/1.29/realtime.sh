#!/bin/bash

set -xe

echo kernel.sched_rt_runtime_us=-1 > /etc/sysctl.d/realtime.conf
sysctl --system
