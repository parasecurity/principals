#include <random>
#include <map>
#include <iostream>
#include <cmath>
#include <boost/math/distributions.hpp>
#include <boost/program_options.hpp>
#include <stdlib.h>
#include <random>
#include <numeric>

extern "C"
{
  #include "tcpclient.h"
}


namespace bpo = boost::program_options;

/*void parse_args(all_args_t* args, int argc, char* argv[]){

  bpo::options_description general("General options");
  general.add_options()
      ("help,h", "Produce help message")
      ("version,v", "Print version information and exit");

    // Set IP section
    ("tcp.coonect.remote_ipv4_addr", bpo::value<string>(&args->ric_agent.remote_ipv4_addr)->default_value("127.0.0.1"), "IPv4 address of the TCP server.");

}*/


int main(int argc, char const *argv[]){
        //all_args_t args = {};

        std::cout << "... Starting Traffic Generation ..." << std::endl;
        //parse_args(&args, argc, argv)


        std::default_random_engine generator;
        typedef std::mt19937 G;
        typedef std::gamma_distribution<double> dist;
        G g;

        //----------------Alexa Model-----------------
        double k_alexa = 0.693;
        double s_alexa = 134915;

        //---------------Ecobee Model----------------
	double k_eco = 28;
	double s_eco = 715.159;

        //---------------Hub Model----------------
        double k_hub = 31;
        double s_hub = 983.76;	

        //------------Smart Plug Mode-------------
        double k_plug = 69;
        double s_plug = 18.013;

        startTCPclient();
	

	int device_count[4] = {50, 300, 300, 350};
	int device_select[4] = {1, 2, 3, 4};
	int device_subtot[4] = {0};
        double send_chance[4] = {100, 64.308, 73.5, 17.547};	
	std::string device_name[4] = {"Alexa", "Ecobee", "SmartHub", "SmartPlug"};
	double k_device[4] = {k_alexa, k_eco, k_hub, k_plug};
	double s_device[4] = {s_alexa, s_eco, s_hub, s_plug};

        int total_out;
	int device;

        int option = 2;


        // Operation mode 1: Select Random Device Each Time 1 in 1 :::::: Alternate Mode: aggregate 1000 samples from same device
        if (option == 1){
                while(1){
                        device = rand() % 4 + 0;
			//for (int i=0; i<=device_count[device]; ++i) {
			dist gamma_rand(k_device[device], s_device[device]);

			if (rand() % 100 < send_chance[device]) device_subtot[device] += ceil(gamma_rand(g));
			else device_subtot[device] += 0;
			//}
			std::cout << "Device chosen is: " << device_name[device] << " sent " << device_subtot[device] << " bytes" << std::endl;
                        sendBytes(device_subtot[device]);
			device_subtot[device] = 0;
                }
        }

        // Operation mode 2: Aggregate 1000 samples from fixed number of devices in 1 
        else if (option == 2){


                while(1){
			for (int j=0; j<=(sizeof(device_count)/sizeof(device_count[0])); j++){	
				device = j;
				for (int i=0; i<=device_count[device]; i++) {
					if (rand() % 100 < send_chance[device]) {
						dist gamma_rand(k_device[device], s_device[device]);
						device_subtot[device] += ceil(gamma_rand(g));
					}
					else device_subtot[device] += 0;
				}
			}	
			total_out = std::accumulate(device_subtot, device_subtot+(sizeof(device_count)/sizeof(device_count[0])), 0);
                        std::cout << "Total: " << total_out << std::endl;
                        sendBytes(total_out);
			total_out = 0;
			int device_subtot[4] = {0};		
		}
                
        }
}
