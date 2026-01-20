#!/bin/bash
chmod +x h5
ps -ef | grep h5 | grep -v grep | awk '{print $2}' | xargs kill -9

nohup ./h5 &


chmod +x consumer
ps -ef | grep consumer | grep -v grep | awk '{print $2}' | xargs kill -9

nohup ./consumer &