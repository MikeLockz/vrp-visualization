import pandas as pd

def analyze_csv(file_path):
    """
    Analyze CSV file to count non-empty 'Outward' columns per row and display sorted summary
    """
    try:
        # Read the CSV file
        print(f"Reading file from: {file_path}")
        df = pd.read_csv(file_path)
        
        # Get columns that contain 'Outward'
        outward_columns = [col for col in df.columns if 'Outward' in str(col)]
        
        print(f"\nFound {len(outward_columns)} columns containing 'Outward'")
        
        # Create a count of non-empty values for each row
        df['outward_count'] = df[outward_columns].notna().sum(axis=1)
        
        # Create summary with Summary column and count
        summary_df = df[['Summary', 'outward_count']].copy()
        
        # Sort by count in descending order
        summary_df = summary_df.sort_values('outward_count', ascending=False)
        
        # Basic statistics
        print(f"\nStatistics for non-empty Outward columns per row:")
        print(f"Average per row: {summary_df['outward_count'].mean():.2f}")
        print(f"Maximum in any row: {summary_df['outward_count'].max()}")
        print(f"Rows with no Outward values: {(summary_df['outward_count'] == 0).sum()}")
        print(f"Rows with at least one Outward value: {(summary_df['outward_count'] > 0).sum()}")
        
        # Print sorted summary
        print("\nSorted summary (tickets with outward links):")
        # Only show rows with counts > 0
        non_zero_summary = summary_df[summary_df['outward_count'] > 0]
        print(non_zero_summary.to_string(index=False))
        
        return summary_df
        
    except Exception as e:
        print(f"Error analyzing CSV: {str(e)}")
        return None

if __name__ == "__main__":
    try:
        file_path = "EXPORT/ptr.csv"
        summary_df = analyze_csv(file_path)
                
    except Exception as e:
        print(f"Error running script: {str(e)}")