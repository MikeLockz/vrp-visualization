import pandas as pd
import plotly.graph_objects as go
from plotly.subplots import make_subplots
from datetime import datetime

def analyze_ticket_trends():
    """
    Create an interactive visualization showing daily ticket creation vs resolution counts
    with moving average trend lines using Plotly
    """
    file_path = "EXPORT/pds.csv"
    
    try:
        # Read the CSV file with low_memory=False to handle mixed types
        df = pd.read_csv(file_path, low_memory=False)
        
        # Convert 'Created' column to datetime
        df['Created'] = pd.to_datetime(df['Created'])
        
        # Create date-only column for grouping
        df['Date'] = df['Created'].dt.date
        
        # Count tickets created per day
        created_counts = df.groupby('Date').size()
        
        # Count resolved tickets per day
        resolved_counts = df[df['Status'].isin(['Done', 'Resolved'])].groupby('Date').size()
        
        # Calculate 7-day moving averages
        created_ma = created_counts.rolling(window=7, center=True).mean()
        resolved_ma = resolved_counts.rolling(window=7, center=True).mean()
        
        # Create the interactive plot
        fig = make_subplots(specs=[[{"secondary_y": True}]])
        
        # Add bars for daily counts
        fig.add_trace(
            go.Bar(
                name="Created",
                x=created_counts.index,
                y=created_counts.values,
                marker_color='rgba(65, 105, 225, 0.5)',
                hovertemplate="Date: %{x}<br>Created: %{y}<extra></extra>"
            )
        )
        
        fig.add_trace(
            go.Bar(
                name="Resolved/Done",
                x=resolved_counts.index,
                y=resolved_counts.values,
                marker_color='rgba(34, 139, 34, 0.5)',
                hovertemplate="Date: %{x}<br>Resolved: %{y}<extra></extra>"
            )
        )
        
        # Add trend lines
        fig.add_trace(
            go.Scatter(
                name="7-day Created Avg",
                x=created_counts.index,
                y=created_ma,
                line=dict(color='rgb(0, 0, 139)', width=2),
                hovertemplate="Date: %{x}<br>7-day Avg Created: %{y:.1f}<extra></extra>"
            )
        )
        
        fig.add_trace(
            go.Scatter(
                name="7-day Resolved Avg",
                x=resolved_counts.index,
                y=resolved_ma,
                line=dict(color='rgb(0, 100, 0)', width=2),
                hovertemplate="Date: %{x}<br>7-day Avg Resolved: %{y:.1f}<extra></extra>"
            )
        )
        
        # Update layout
        fig.update_layout(
            title="Daily Ticket Creation vs Resolution",
            title_x=0.5,
            xaxis_title="Date",
            yaxis_title="Number of Tickets",
            barmode='overlay',
            hovermode='x unified',
            template='plotly_white',
            showlegend=True,
            legend=dict(
                yanchor="top",
                y=0.99,
                xanchor="left",
                x=0.01
            ),
            # Add range slider
            xaxis=dict(
                rangeslider=dict(visible=True),
                type="date"
            )
        )
        
        # Save as interactive HTML
        fig.write_html("ticket_trends_interactive.html")
        print("\nInteractive analysis saved as ticket_trends_interactive.html")
        
        # Print summary statistics
        print("\nSummary Statistics:")
        print(f"Total tickets created: {len(df):,}")
        print(f"Total tickets resolved: {len(df[df['Status'].isin(['Done', 'Resolved'])]):,}")
        print("\nDaily averages:")
        print(f"Average daily created: {created_counts.mean():.1f}")
        print(f"Average daily resolved: {resolved_counts.mean():.1f}")
        
        # Calculate and print backlog trend
        backlog = len(df) - len(df[df['Status'].isin(['Done', 'Resolved'])])
        print(f"\nCurrent backlog: {backlog:,} tickets")
        
        # Calculate average resolution rate
        days_span = (df['Created'].max() - df['Created'].min()).days
        if days_span > 0:
            resolution_rate = len(df[df['Status'].isin(['Done', 'Resolved'])]) / days_span
            print(f"Average resolution rate: {resolution_rate:.1f} tickets per day")
            if backlog > 0 and resolution_rate > 0:
                estimated_days_to_clear = backlog / resolution_rate
                print(f"Estimated days to clear backlog at current rate: {estimated_days_to_clear:.1f} days")
                print(f"Estimated backlog clearance date: {df['Created'].max().date() + pd.Timedelta(days=estimated_days_to_clear)}")
        
    except FileNotFoundError:
        print("Error: Could not find file at", file_path)
    except Exception as e:
        print("Error processing file:", str(e))

# Run the analysis
if __name__ == "__main__":
    analyze_ticket_trends()