package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	exam "github.com/priority-infra/go_exam/internal/protos"
	pb "github.com/priority-infra/go_exam/internal/protos" // Add this line to import the pb package
	"github.com/priority-infra/go_exam/internal/data" // Add this line to import the data package
)

// Version represents the version of the service
var Version string = "development"

const Service = "exam"

// ListenHost Host and port to listen
var ListenHost string = "0.0.0.0:8000"

// Server this is a main object of the grpc server
var Server *grpc.Server

type server struct{}

func (s *server) GetManagerProjects(_ *pb.Empty, stream pb.Exam_GetManagerProjectsServer) error {
	managerProjectCount := make(map[int64]int64)

	// Step 1: Count projects for each manager
	for _, project := range data.Projects {
		for _, department := range data.Departments {
			if project.DepartmentID == department.ID {
				managerProjectCount[department.ManagerID]++
			}
		}
	}

	// Step 2: Filter managers with more than 1 project
	var managerList []pb.ManagerProjects
	for managerID, count := range managerProjectCount {
		if count > 1 {
			for _, employee := range data.Employees {
				if employee.ID == managerID {
					managerList = append(managerList, pb.ManagerProjects{
						ManagerName:  employee.Name,
						ProjectCount: count,
					})
				}
			}
		}
	}

	// Step 3: Sort by project count in descending order
	sort.Slice(managerList, func(i, j int) bool {
		return managerList[i].ProjectCount > managerList[j].ProjectCount
	})

	// Stream the results to the client
	for _, manager := range managerList {
		if err := stream.Send(&manager); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	lis, err := net.Listen("tcp", ListenHost)
	if err != nil {
	Server = grpc.NewServer()
		exam.RegisterExamServer(Server, &server{})
	Server = grpc.NewServer()
	exam.RegisterExamServer(Server, NewExamServer())

	log.Printf("starting to listen of %s, press ctrl+c to stop", ListenHost)
	go closeHandler()
	Server.Serve(lis)
}

func closeHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	Server.GracefulStop()
	log.Printf("exiting gracefully")
}
