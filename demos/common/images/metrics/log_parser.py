#!/bin/python3

try:
    flow = open("flow-control.log")
    antrea = open("antrea.log")
except:
    print("Error opening logs")

f = [x for x in flow]
a = [x for x in antrea]

flow_set = {'Forwarding', 'Done'}
antrea_set = {'Listening', 'Received', 'Applied'}

f = [x.split(' ') for x in f]
f = [(float(x[1]), x[2]) for x in f]

a = [x.split(' ') for x in a]
a = [(float(x[1]), x[2]) for x in a]

f= f + a
f.sort(key=lambda a: a[0])

print("| {0:12}| {1:23}| {2:23}| {3:23}|".format("Action", "Timestamp (unix)", "dts (ms from start)", "dt (ms from prev act)"))
for i in f[1:]:
    if i[1] == 'Forwarding':
        print("|========================================================================================|")
        t_start = i[0]
        t_prev = i[0]
        dt = 0
        dts = 0
    else:
        dts = (i[0] - t_start)*1000 
        dt = (i[0] - t_prev)*1000
        t_prev = i[0]

    print("| {0:12}| t ={1:20}| +{2:20}ms| +{3:20}ms|".format(i[1], i[0], dts, dt))

