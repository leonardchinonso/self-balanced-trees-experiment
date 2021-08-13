package main

import (
	"context"
	model "es/model"
	"fmt"
	"log"
	"os"
	"runtime"
)

const (
	pathToGit = "C:\\Program Files\\Git\\cmd\\git.exe"
)

type (
	erigonService struct {
		model.ErigonServiceServer
	}
)

func (es *erigonService) validateRequest(request *model.BranchRequest) (string, error) {
	if es == nil {
		return "", fmt.Errorf("BuildFrom called on nil object")
	}

	if request == nil {
		return "", fmt.Errorf("BuildFrom called with invalid parameter value nil")
	}

	branchName := request.BranchName
	if len(branchName) == 0 {
		return "", fmt.Errorf("provide a valid branch name")
	}

	return branchName, nil
}

func (es *erigonService) checkOutBranchAndRunErigon(branchName string) {
	log.Printf("Checking out branch: %v", branchName)

	if runtime.GOOS != "windows" {
		fmt.Println("Can only execute on Windows machines")
	} else {
		es.execute(branchName)
	}
}

func (es *erigonService) BuildFrom(ctx context.Context, request *model.BranchRequest) (*model.ErigonResponse, error) {
	branchName, err := es.validateRequest(request)
	if err != nil {
		return nil, err
	}

	// Checkout branch by branch name and build Erigon
	es.checkOutBranchAndRunErigon(branchName)

	// Initialise result after building erigon
	result := &model.ErigonResponse{}
	result.Msg = "Done with the process"

	return result, nil
}

func (es *erigonService) execute(branchName string) {
	checkOutCommand := []string {pathToGit, "checkout", branchName}

	var procAttr os.ProcAttr
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	
	if _, err := os.StartProcess(pathToGit, checkOutCommand, &procAttr); err != nil {
		log.Fatal("error starting process, error: ", err)
	}

	procAttr.Dir = "C:\\Users\\USER\\Projects\\Starkware\\gRPCBuildErigon\\Makefile"
	makeErigonCommand := []string {"make", "erigon"}
	if _, err := os.StartProcess(makeErigonCommand[0], makeErigonCommand, &procAttr); err != nil {
		log.Fatal("can not make erigon, error: ", err)
	}
	
	makeRPCDaemonCommand := []string {"make", "rpcdaemon"}
	if _, err := os.StartProcess(makeRPCDaemonCommand[0], makeRPCDaemonCommand, &procAttr); err != nil {
		log.Fatal("can not make rpcdaemon, error: ", err)
	}
}
