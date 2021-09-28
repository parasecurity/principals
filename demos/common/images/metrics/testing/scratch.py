#!/usr/bin/python3
import time
import sys
import argparse
import os

# old version
# user for metric_server.go (first version)
#
def analyze_v1(result_file, log_file_name):
    log_file = open(log_file_name, "r")
    lines = log_file.readlines()
    log_file.close()
    f = [x.split(' ') for x in lines]
        
    s1, c1 = (0, 0) # total client server delay
    s2, c2 = (0, 0) # formated logs client server delay
    s3, c3 = (0, 0) # not formated logs client server delay

    s4, c4 = (0, 0) # formated client delay
    s5, c5 = (0, 0) # not formated client delay
    c6 = 0          # corrupted logs

    for l in f:
        if len(l) < 5: 
            print(l)
            continue
        try:
            if len(l) < 5: print("should not run")
            if l[2][0] != '2':
                print(l)
            else:
                print(l)

                time1 = str_to_microsec(l[0]+' '+l[1])
                if time1 == 0:
                    c6 += 1
                    continue

                time2 = str_to_microsec(l[2]+' '+l[3])
                if time2 == 0:
                    c6 += 1
                    continue

                dt = time1 - time2
                #print("{} {} dt {}".format(time1, time2, dt))
                s1, c1 = (s1 + dt, c1 + 1)

                if len(l) == 6:
                    s2, c2 = (s2 + dt, c2 + 1)
                    if l[5][0] == 'f': 
                        s4, c4 = (s4 + int(l[4]), c4 + 1)
                    elif l[5][0] == 'r': 
                        s5, c5 = (s5 + int(l[4]), c5 + 1)

                else:
                    s3, c3 = (s3 + dt, c3 + 1)
        except:
            c6 += 1


            
    print("================ Client - Server delay =================")
    print("total: "+str(c1)+", formated: "+str(c2)+", not formated: "+str(c3))
    print("average: "+str(s1/c1))
    print("formated: "+str(s2/c2))
    print("not formated: "+str(s3/c3))
    print("================ Client logging delay =================")
    print("formated: "+str((s4/c4)/1000))
    print("not formated: "+str((s5/c5)/1000))
    print("corrupted log entries: "+str(c6))
    return ((s4/c4)/1000), ((s5/c5)/1000), c6

###################################################################################################
#
#
#
#
###################################################################################################

def str_to_microsec(time_string: str) -> int:
    time_format = "%Y/%m/%d %H:%M:%S"
    temp = time_string.split('.')
    try:
        tm = time.mktime(
                time.strptime(temp[0]+"UTC", time_format+"%Z")
                )
    except:
        print("Corrupted log entry")
        return 0
        

    return int(tm)*pow(10, 6) + int(temp[1])

def analyze_v2(result_file, log_file_name):
    log_file = open(log_file_name, "r")
    lines = log_file.readlines()
    log_file.close()
    f = [x.split(' ') for x in lines]
        
    s1, c1 = (0, 0) # total client server delay
    s2, c2 = (0, 0) # formated logs client server delay
    s3, c3 = (0, 0) # not formated logs client server delay

    s4, c4 = (0, 0) # formated client delay
    s5, c5 = (0, 0) # not formated client delay
    c6 = 0          # corrupted logs

    for l in f:
        if len(l) < 4: 
            print(l)
            continue
        try:
            if len(l) < 4: print("should not run")
            if l[1][0] != '2':
                print(l)
            else:
                print(l)


                if len(l) == 5:
                    # s2, c2 = (s2 + dt, c2 + 1)
                    if l[4][0] == 'f': 
                        s4, c4 = (s4 + int(l[3]), c4 + 1)
                    elif l[4][0] == 'r': 
                        s5, c5 = (s5 + int(l[3]), c5 + 1)
        except:
            c6 += 1

    # print("================ Client - Server delay =================")
    # print("total: "+str(c1)+", formated: "+str(c2)+", not formated: "+str(c3))
    # print("average: "+str(s1/c1))
    # print("formated: "+str(s2/c2))
    # print("not formated: "+str(s3/c3))
    print("================ Client logging delay =================")
    print("formated: microseconds"+str((s4/c4)/1000))
    print("not formated: microseconds"+str((s5/c5)/1000))
    print("corrupted log entries: "+str(c6))
    return ((s4/c4)/1000), ((s5/c5)/1000), c6

if __name__ == "__main__":

    parser = argparse.ArgumentParser(description="Basic testing suite")
    parser.add_argument("-c", "--client", help="Go client tester", required=True)
    parser.add_argument("-o", "--output", help="File to save results", required=True)
    parser.add_argument("-l", "--logfile", help="Log file", required=True)
    parser.add_argument("-w", "--workload", help="Workload value for go testers", required=True)
    parser.add_argument("-t", "--tests", help="Times test will be run. Default 1000", required=False)

    args = parser.parse_args()

    if args.tests == None:
        test_no = 1000
    else:
        test_no = args.tests

    fm_total, nfm_total, cor_total = (0, 0, 0)
    res_file = open(args.output, "w")
    os.dup2(res_file.fileno(), sys.stdout.fileno())

    for i in range(0,test_no):
        print("test no: {}".format(i), file=sys.stderr)
        if i%100 == 0:
            os.system("notify-send \"Testing {}\" \"progress: running test {}.\nworkload: {}\"".format(args.client, i, args.workload))
        os.system("echo > {}".format(args.logfile))
        os.system("go run {} -l {}".format(args.client, args.workload))
        time.sleep(0.1)
        fm, nfm, cor = analyze_v1(args.output, args.logfile)
        # fm, nfm, cor = analyze_v2(args.output, args.logfile)
        fm_total += fm
        nfm_total += nfm
        cor_total += cor

    print("=======================================================")
    print("                ======================                 ")
    print("testing {} times".format(test_no))
    print("formated: "+str(fm_total/test_no)+" microseconds")
    print("not formated: "+str(nfm_total/test_no)+" microseconds")
    print("corrupted log entries: "+str(cor_total/test_no)+" microseconds")



