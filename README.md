# Job Coordinator

## Purpose

The purpose of this project supply the foundation of creating a workflow consisting of these workers:
- Scheduler - determine on which machine to run a job that is ready
//- Export - responsible for prepping the local files (could create a script to run locally)
//- CopyTo - copy files to remote machine that will be required for the job
//- Starter - starts the command remotely, once the job started running its job is done
//- Monitor - monitors the remote command to determine when it is done
//- CopyBack - copies the results of the job back
//- PostProcessing - process the results to determine if it succeeded/failed
- Reporting - will probably be responsible for summarizing the post-processed results
- Health - can monitor health of workers and create reports or send emails
- Cleanup worker - could be responsible for cleanup up "temp export" and "remote job" folders that are done