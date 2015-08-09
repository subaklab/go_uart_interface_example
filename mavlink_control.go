package main

import (
	"fmt"
)

func top() {
	//인자를 받아서 처리
	// Default input arguments
	char *uart_name = (char*)"/dev/ttyUSB0";
	int baudrate = 57600;

	// do the parse, will throw an int if it fails
	parse_commandline(argc, argv, uart_name, baudrate);

	Serial_Port serial_port(uart_name, baudrate);

	Autopilot_Interface autopilot_interface(&serial_port);

	serial_port_quit         = &serial_port;
	autopilot_interface_quit = &autopilot_interface;
	signal(SIGINT,quit_handler);

	/*
	 * Start the port and autopilot_interface
	 * This is where the port is opened, and read and write threads are started.
	 */
	serial_port.start();
	autopilot_interface.start();

	commands(autopilot_interface);

	autopilot_interface.stop();
	serial_port.stop();


}

func commands(api *Autopilot_Interface){
	api.enable_offboard_control();
	usleep(100); // give some time to let it sink in

	fmt.Println("SEND OFFBOARD COMMANDS\n");

	// initialize command data strtuctures
	mavlink_set_position_target_local_ned_t sp;
	mavlink_set_position_target_local_ned_t ip = api.initial_position;

	// Example 2 - Set Position
	set_position( ip.x - 5.0 , // [m]
			 	   ip.y - 5.0 , // [m]
				   ip.z       , // [m]
				   sp         );


	// Example 1.2 - Append Yaw Command
	set_yaw( ip.yaw , // [rad]
			 sp     );

	// SEND THE COMMAND
	api.update_setpoint(sp);
	// NOW pixhawk will try to move

	// Wait for 8 seconds, check position
	for (int i=0; i < 8; i++)
	{
		mavlink_local_position_ned_t pos = api.current_messages.local_position_ned;
		printf("%i CURRENT POSITION XYZ = [ % .4f , % .4f , % .4f ] \n", i, pos.x, pos.y, pos.z);
		sleep(1);
	}

	fmt.Println("\n");

	api.disable_offboard_control();

	fmt.Println("READ SOME MESSAGES \n");

	Mavlink_Messages messages = api.current_messages;

	// local position in ned frame
	mavlink_local_position_ned_t pos = messages.local_position_ned;
	fmt.Println("Got message LOCAL_POSITION_NED (spec: https://pixhawk.ethz.ch/mavlink/#LOCAL_POSITION_NED)\n");
	fmt.Println("    pos  (NED):  %f %f %f (m)\n", pos.x, pos.y, pos.z );

	// hires imu
	mavlink_highres_imu_t imu = messages.highres_imu;
	fmt.Println("Got message HIGHRES_IMU (spec: https://pixhawk.ethz.ch/mavlink/#HIGHRES_IMU)\n");
	fmt.Println("    ap time:     %lu \n", imu.time_usec);
	fmt.Println("    acc  (NED):  % f % f % f (m/s^2)\n", imu.xacc , imu.yacc , imu.zacc );
	fmt.Println("    gyro (NED):  % f % f % f (rad/s)\n", imu.xgyro, imu.ygyro, imu.zgyro);
	fmt.Println("    mag  (NED):  % f % f % f (Ga)\n"   , imu.xmag , imu.ymag , imu.zmag );
	fmt.Println("    baro:        %f (mBar) \n"  , imu.abs_pressure);
	fmt.Println("    altitude:    %f (m) \n"     , imu.pressure_alt);
	fmt.Println("    temperature: %f C \n"       , imu.temperature );
}
func parse_commandline(string uart_name, int baudrate){
	// string for command line usage
	const char *commandline_usage = "usage: mavlink_serial -d <devicename> -b <baudrate>";

	// Read input arguments
	for (int i = 1; i < argc; i++) { // argv[0] is "mavlink"

		// Help
		if (strcmp(argv[i], "-h") == 0 || strcmp(argv[i], "--help") == 0) {
			printf("%s\n",commandline_usage);
			throw EXIT_FAILURE;
		}

		// UART device ID
		if (strcmp(argv[i], "-d") == 0 || strcmp(argv[i], "--device") == 0) {
			if (argc > i + 1) {
				uart_name = argv[i + 1];

			} else {
				printf("%s\n",commandline_usage);
				throw EXIT_FAILURE;
			}
		}

		// Baud rate
		if (strcmp(argv[i], "-b") == 0 || strcmp(argv[i], "--baud") == 0) {
			if (argc > i + 1) {
				baudrate = atoi(argv[i + 1]);

			} else {
				printf("%s\n",commandline_usage);
				throw EXIT_FAILURE;
			}
		}

	}
}

func quit_handler(sing int){
	fmt.Println("");
	fmt.Println("TERMINATING AT USER REQUEST\n");
	fmt.Println("\n");

	// autopilot interface
	try {
		autopilot_interface_quit->handle_quit(sig);
	}
	catch (int error){}

	// serial port
	try {
		serial_port_quit->handle_quit(sig);
	}
	catch (int error){}

	// end program here
	exit(0);

}

//quit handler
var autopilot_interface_quit *Autopilot_Interface
serial_port_quit *Serial_Port


func main() {
	try
	{
		int result = top(argc,argv);
		return result;
	}

	catch ( int error )
	{
		fprintf(stderr,"mavlink_control threw exception %i \n" , error);
		return error;
	}
}
