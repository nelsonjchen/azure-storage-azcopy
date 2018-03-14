// Copyright © 2017 Microsoft <wastore@microsoft.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"github.com/Azure/azure-storage-azcopy/common"
	"github.com/spf13/cobra"
	"math"
	"encoding/json"
)

type ListReq struct {
	JobID    common.JobID
	StatusOf common.TransferStatus
}

type ListResponse struct {
	ErrorMsg string
}

type ListExistingJobResponse struct {
}

type ListProgressSummaryResponse struct {
}

func init() {
	commandLineInput := common.ListRequest{}

	// lsCmd represents the ls command
	lsCmd := &cobra.Command{
		Use:        "list",
		Aliases:    []string{"ls"},
		SuggestFor: []string{"lst", "lt", "ist"}, //TODO why does message appear twice on the console
		Short:      "list(ls) lists the specifics of an existing job.",
		Long: `list(ls) lists the specifics of an existing job. The most common cases are:
  - lists all the existing jobs.
  - lists all the part numbers of an existing jobs.
  - lists all the transfers of an existing job.`,
		Args: func(cmd *cobra.Command, args []string) error {
			// the list command does not require necessarily to have an argument
			/*
			* list -- lists all the existing jobs
			* list jobId -- lists the progress summary of the job for given jobId
			* list jobId --with-status -- lists all the transfers of an existing job which has the given status
			 */

			// if there is more than one argument passed, then it is taken as a jobId
			if len(args) > 0 {
				jobId, err := common.ParseJobID(args[0])
				if err != nil {
					fmt.Print("invalid job Id given ", args[0])
					return nil
				}
				commandLineInput.JobId = jobId
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			HandleListCommand(commandLineInput)
		},
	}

	rootCmd.AddCommand(lsCmd)

	// define the flags relevant to the ls command

	// filters
	lsCmd.PersistentFlags().StringVar(&commandLineInput.OfStatus, "with-status", "", "Filter: list transfers of job only with this status")
}

// handles the list command
// dispatches the list order to theZiyi Wang storage engine
func HandleListCommand(commandLineInput common.ListRequest) {

	// check whether ofstatus transfer status is valid or not
	if commandLineInput.OfStatus != "" &&
		common.TransferStatusStringToCode(commandLineInput.OfStatus) == math.MaxUint32 {
		fmt.Println("invalid transfer status passed. Please provide the correct transfer status flag")
		return
	}

	var response []byte

	if commandLineInput.JobId == common.EmptyJobId {
		response, _ = common.Rpc("listJobs", commandLineInput)
	} else if commandLineInput.OfStatus == "" {
		response, _ = common.Rpc("listJobProgressSummary", commandLineInput)
	} else {
		response, _ = common.Rpc("listJobTransfers", commandLineInput)
	}

	// list Order command requested the list of existing jobs
	if commandLineInput.JobId == common.EmptyJobId {
		PrintExistingJobIds(response)
	} else if commandLineInput.OfStatus == "" { //list Order command requested the progress summary of an existing job
		PrintJobProgressSummary(response)
	} else { //list Order command requested the list of specific transfer of an existing job
		PrintJobTransfers(response)
	}
}

// PrintExistingJobIds prints the response of listOrder command when listOrder command requested the list of existing jobs
func PrintExistingJobIds(data []byte) {
	var listJobResponse common.ListJobsResponse
	err := json.Unmarshal(data, &listJobResponse)
	if err != nil {
		panic(err)
	}
	if listJobResponse.Errormessage != "" {
		fmt.Println(fmt.Sprintf("request failed with following error message %s", listJobResponse.Errormessage))
		return
	}

	fmt.Println("Existing Jobs ")
	for index := 0; index < len(listJobResponse.JobIds); index++ {
		fmt.Println(listJobResponse.JobIds[index].String())
	}
}

// PrintJobTransfers prints the response of listOrder command when list Order command requested the list of specific transfer of an existing job
func PrintJobTransfers(data []byte) {
	var listTransfersResponse common.ListJobTransfersResponse
	err := json.Unmarshal(data, &listTransfersResponse)
	if err != nil {
		panic(err)
	}
	if listTransfersResponse.ErrorMessage != "" {
		fmt.Println(fmt.Sprintf("request failed with following message %s", listTransfersResponse.ErrorMessage))
		return
	}

	fmt.Println(fmt.Sprintf("----------- Transfers for JobId %s -----------", listTransfersResponse.JobId))
	for index := 0; index < len(listTransfersResponse.Details); index++ {
		fmt.Println(fmt.Sprintf("transfer--> source: %s destination: %s status %s", listTransfersResponse.Details[index].Src, listTransfersResponse.Details[index].Dst,
			listTransfersResponse.Details[index].TransferStatus))
	}
}

// PrintJobProgressSummary prints the response of listOrder command when listOrder command requested the progress summary of an existing job
func PrintJobProgressSummary(summaryData []byte) {
	var summary common.ListJobSummaryResponse
	err := json.Unmarshal(summaryData, &summary)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling the progress summary. Failed with error %s", err.Error()))
		return
	}
	if summary.ErrorMessage != "" {
		fmt.Println(fmt.Sprintf("list progress summary of job failed because %s", summary.ErrorMessage))
		return
	}
	fmt.Println(fmt.Sprintf("--------------- Progress Summary for Job %s ---------------", summary.JobId))
	fmt.Println("Total Number of Transfer ", summary.TotalNumberOfTransfers)
	fmt.Println("Total Number of Transfer Completed ", summary.TotalNumberofTransferCompleted)
	fmt.Println("Total Number of Transfer Failed ", summary.TotalNumberofFailedTransfer)
	fmt.Println("Has the final part been ordered ", summary.CompleteJobOrdered)
	fmt.Println("Progress of Job in terms of Perecentage ", summary.PercentageProgress)
	for index := 0; index < len(summary.FailedTransfers); index++ {
		message := fmt.Sprintf("transfer-%d	source: %s	destination: %s", index, summary.FailedTransfers[index].Src, summary.FailedTransfers[index].Dst)
		fmt.Println(message)
	}
}
