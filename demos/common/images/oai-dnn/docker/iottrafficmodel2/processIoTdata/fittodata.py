import numpy as np
import scipy
import scipy.stats
import matplotlib.pyplot as plt
from distfit import distfit

datain = np.loadtxt("processedin.txt", dtype=int)
dataout = np.loadtxt("processedin.txt", dtype=int)

datain=datain
dataout=dataout

nzdatain = np.nonzero(datain)[0] 
nzsize = nzdatain.size
print(nzdatain)

tsize = datain.size
datain = datain[datain != 0]
print(datain)

print("Total Active Percentage")
print(nzsize/tsize)


#distout = distfit()
#distout.fit_transform(dataout)
#print("Outgoing data distribution")
#print(distout.summary)

distin = distfit()
distin.fit_transform(datain)
print("Incoming data distribution")
print(distin.summary)

size = datain.size
datapoints = size
x = np.arange(datapoints)/10

#plt.xlabel('Time (sec)')
#plt.ylabel('Data sent (KB)')
#plt.plot(x,datain[0:datapoints])
##plt.ylim([0,400])
#plt.savefig('samhubtraffic_unscaled.png')

#h = plt.hist(datain, bins=range(30))

dist_names = ['gamma', 'beta', 'rayleigh', 'norm', 'pareto']

for dist_name in dist_names:
    dist = getattr(scipy.stats, dist_name)
    params = dist.fit(datain)
    arg = params[:-2]
    loc = params[-2]
    scale = params[-1]
    #if arg:
    #    pdf_fitted = dist.pdf(x, *arg, loc=loc, scale=scale) * size
    #else:
    #    pdf_fitted = dist.pdf(x, loc=loc, scale=scale) * size
    #plt.plot(pdf_fitted, label=dist_name)
    #plt.xlim(0,47)
#plt.legend(loc='upper right')
#plt.show()

print(arg)
print(loc)
print(scale)
