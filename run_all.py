#!/usr/bin/env python3
import subprocess
import sys
import os
from datetime import datetime

def run_scripts():
    # List of scripts to run in order
    scripts = [
        "download_csv-jira_export.py",
        "process_csv-pds-histogram.py",
        "process_csv-ptr-count_of_linked_tickets.py",
        "process_csv-ptr-pds-linked-ticket-created-dates.py",
        "process_csv-pds-labels.py"
    ]
    
    # Create logs directory if it doesn't exist
    if not os.path.exists('logs'):
        os.makedirs('logs')
    
    # Log file with timestamp
    log_file = f"logs/script_run_{datetime.now().strftime('%Y%m%d_%H%M%S')}.log"
    
    with open(log_file, 'w') as log:
        for script in scripts:
            print(f"\nRunning {script}...")
            log.write(f"\n{'='*50}\nRunning {script} at {datetime.now()}\n{'='*50}\n")
            
            try:
                # Run the script and capture output
                result = subprocess.run(
                    [sys.executable, script],
                    capture_output=True,
                    text=True
                )
                
                # Log stdout and stderr
                log.write(f"STDOUT:\n{result.stdout}\n")
                log.write(f"STDERR:\n{result.stderr}\n")
                
                if result.returncode != 0:
                    print(f"Error running {script}! Check {log_file} for details.")
                    log.write(f"Script failed with return code {result.returncode}\n")
                else:
                    print(f"Successfully completed {script}")
                    log.write("Script completed successfully\n")
                    
            except Exception as e:
                error_msg = f"Failed to run {script}: {str(e)}"
                print(error_msg)
                log.write(f"{error_msg}\n")

if __name__ == "__main__":
    run_scripts()