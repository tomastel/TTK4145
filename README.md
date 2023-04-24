# [TTK4145: Real Time Programming](https://www.ntnu.edu/studies/courses/TTK4145/#tab=omEmnet)

*Project in collaboration with [Khuong Huynh](https://github.com/Khuongh) and [Maiken Pedersen](https://github.com/maikenpedersen).*


This repository contains all files of our result for the elevator project, as well as some of the mandatory exercises, in the course [TTK4145 Real Time Programming](https://www.ntnu.edu/studies/courses/TTK4145/#tab=omEmnet) taught at NTNU. The course is taught as part of the master's programme [Cybernetics and Robotics](https://www.ntnu.edu/studies/mttk).

The elevator project aims to teach the students about software design, real-time and distributed systems through designing and writing a distributed program for an elevator system consisting of n elevators and m floors. The students are free to decide how they wish to design the system, including distributed system type (most students choose P2P or master-slave), communication protocols (TCP, UDP, both or other) and programming language. We chose to design a P2P-system using UDP and writing the program in [Go](https://go.dev/).

Several requirements must be met in order to complete the project. The system should distribute orders efficiently so that an elevator arrives at the given floor as fast as possible, and the network should handle simultaneuous calls, nodes disconnecting and reconnecting, high packet loss and obstruction by the door sensor. Most importantly; **no orders are to be lost**.

Feel free to read the [project requirements](https://github.com/tomastel/TTK4145/blob/main/Elevator%20project/Project%20requirements.md) for more detailed functionality.
