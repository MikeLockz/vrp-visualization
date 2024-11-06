from jiraone import LOGIN, issue_export
import json
import os
from datetime import datetime, timedelta

def get_full_export_path(filename):
    """
    Get the full path where jiraone will save the export file
    
    Args:
        filename (str): Base filename
        
    Returns:
        str: Full path to the export file
    """
    export_dir = os.path.join(os.getcwd(), "EXPORT")
    base_name = filename.replace('.csv', '')
    return os.path.join(export_dir, f"{base_name}.csv")

def is_file_fresh(filepath, max_age_days=1):
    """
    Check if file exists and is newer than max_age_days
    
    Args:
        filepath (str): Path to the file to check
        max_age_days (int): Maximum age in days to consider file fresh
        
    Returns:
        bool: True if file exists and is fresh, False otherwise
    """
    if not os.path.exists(filepath):
        return False
        
    file_time = datetime.fromtimestamp(os.path.getmtime(filepath))
    age = datetime.now() - file_time
    
    return age < timedelta(days=max_age_days)

def export_jira_data(jql, export_file):
    """
    Helper function to export Jira data based on JQL query if needed
    
    Args:
        jql (str): JQL query string
        export_file (str): Output file path
    """
    full_path = get_full_export_path(export_file)
    
    if is_file_fresh(full_path):
        print(f"Skipping {export_file} - file exists and is less than 1 day old")
        return

    print(f"Downloading fresh data to {export_file}...")
    try:
        issue_export(
            jql=jql,
            final_file=export_file,
            encoding='utf-8'
        )
        print(f"Successfully exported data to {export_file}")
    except Exception as e:
        print(f"Error exporting data for query '{jql}': {str(e)}")

# Load config
file = "config.json"
config = json.load(open(file))

# Convert config keys to match jiraone expectations
jira_config = {
    "user": config['email'],
    "password": config['token'],
    "url": config['url']
}

# Login to Jira
LOGIN(**jira_config)

# Define JQL queries for different boards
# You can modify these queries based on your actual projects
board1_jql = 'project = PDS ORDER BY "Time to resolution" ASC'
board2_jql = 'project = "PTR" ORDER BY created DESC'  # Modify this to match your actual project

# Define export file paths
board1_export = "pds.csv"
board2_export = "ptr.csv"

# Create EXPORT directory if it doesn't exist
export_dir = os.path.join(os.getcwd(), "EXPORT")
os.makedirs(export_dir, exist_ok=True)

# Export data from both boards if needed
print("\nExporting board 1...")
export_jira_data(board1_jql, board1_export)

print("\nExporting board 2...")
export_jira_data(board2_jql, board2_export)

print("\nExport process complete!")