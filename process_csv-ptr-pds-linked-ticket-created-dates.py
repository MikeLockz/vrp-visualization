import pandas as pd
import plotly.graph_objects as go
from datetime import datetime, timedelta
from collections import defaultdict

def create_ptr_visualization(ptr_path, pds_path):
    """
    Create an interactive visualization of PTR tickets and their linked PDS tickets,
    combining low-count tickets into an 'Other' category
    """
    # Read the CSV files
    ptr_df = pd.read_csv(ptr_path)
    pds_df = pd.read_csv(pds_path)
    
    # Convert Created column to datetime
    pds_df['Created'] = pd.to_datetime(pds_df['Created'])
    
    # Get columns that contain 'Outward'
    outward_columns = [col for col in ptr_df.columns if 'Outward' in str(col)]
    
    # Get all weekly ranges since January 2024
    start_date = datetime(2024, 1, 1)
    current_date = datetime.now()
    
    # Create weekly ranges
    weeks = []
    week_start = start_date
    while week_start <= current_date:
        week_end = week_start + timedelta(days=6)
        weeks.append((week_start, week_end))
        week_start = week_end + timedelta(days=1)
    
    # Structure to store data for visualization
    ptr_data = defaultdict(lambda: {
        'total_links': 0,
        'weekly_counts': defaultdict(int),
        'summary': ''
    })
    
    # Process the data
    for _, ptr_row in ptr_df.iterrows():
        ptr_key = ptr_row['Issue key']
        ptr_data[ptr_key]['summary'] = ptr_row['Summary']
        
        # Count total linked tickets
        for col in outward_columns:
            if pd.notna(ptr_row[col]):
                ptr_data[ptr_key]['total_links'] += 1
                pds_key = ptr_row[col]
                matching_pds = pds_df[pds_df['Issue key'] == pds_key]
                
                if not matching_pds.empty:
                    created_date = matching_pds.iloc[0]['Created']
                    # Find which week this ticket belongs to
                    for week_start, week_end in weeks:
                        if week_start <= created_date <= week_end:
                            week_key = week_start.strftime('%Y-%m-%d')
                            ptr_data[ptr_key]['weekly_counts'][week_key] += 1
    
    # Create the visualization
    fig = go.Figure()
    
    # High contrast color palette for distinct PTRs
    colors = [
        '#003F5C',  # Dark Blue
        '#58508D',  # Purple
        '#BC5090',  # Pink
        '#FF6361',  # Coral
        '#FFA600',  # Orange
        '#005F73',  # Teal
        '#0A9396',  # Light Teal
        '#94D2BD',  # Mint
        '#E9D8A6',  # Sand
        '#EE9B00',  # Gold
        '#CA6702',  # Burnt Orange
        '#BB3E03',  # Red Orange
        '#AE2012',  # Red
        '#9B2226',  # Dark Red
        '#005377',  # Ocean Blue
        '#0088BC',  # Bright Blue
        '#2F9E44',  # Green
        '#486B00',  # Olive
        '#1A472A',  # Forest Green
        '#2B50AA'   # Royal Blue
    ]
    
    # Get all unique weeks
    all_weeks = sorted(set(
        week 
        for data in ptr_data.values() 
        for week in data['weekly_counts'].keys()
    ))
    
    # Assign fixed colors to PTRs based on total ticket count
    ptr_colors = {}
    sorted_ptrs = sorted(
        ptr_data.items(),
        key=lambda x: sum(x[1]['weekly_counts'].values()),
        reverse=True
    )
    for i, (ptr_key, _) in enumerate(sorted_ptrs):
        ptr_colors[ptr_key] = colors[i % len(colors)]
    
    # For each week, create bars ordered by that week's count
    for week in all_weeks:
        # Split PTRs into main and other categories
        week_ptrs = []
        other_count = 0
        other_details = []
        
        for ptr_key, data in ptr_data.items():
            count = data['weekly_counts'].get(week, 0)
            if count >= 2:
                week_ptrs.append((ptr_key, data))
            elif count > 0:
                other_count += count
                other_details.append(f"{ptr_key} ({count})")
        
        # Sort PTRs by count for this week
        week_ptrs.sort(key=lambda x: x[1]['weekly_counts'][week], reverse=True)
        
        # Add traces for each PTR for this week
        for ptr_key, data in week_ptrs:
            fig.add_trace(go.Bar(
                name=f"{ptr_key}: {data['summary'][:30]}...",
                x=[week],
                y=[data['weekly_counts'][week]],
                marker_color=ptr_colors[ptr_key],
                hovertemplate=(
                    f"<b>{ptr_key}</b><br>" +
                    f"{data['summary']}<br>" +
                    "<b>New tickets this week: %{y}</b><br>" +
                    f"Total linked tickets: {data['total_links']}<br>" +
                    "Week: %{x}<br>" +
                    "<extra></extra>"
                ),
                showlegend=False
            ))
        
        # Add other category if there are any low-count tickets
        if other_count > 0:
            details_text = "<br>".join(sorted(other_details))
            fig.add_trace(go.Bar(
                name="Other PTRs",
                x=[week],
                y=[other_count],
                marker_color='#CCCCCC',  # Gray color for Other
                hovertemplate=(
                    "<b>Other PTRs</b><br>" +
                    f"Combined tickets this week: {other_count}<br>" +
                    f"Included PTRs:<br>{details_text}<br>" +
                    "Week: %{x}<br>" +
                    "<extra></extra>"
                ),
                showlegend=False
            ))
    
    # Update layout
    fig.update_layout(
        title={
            'text': "Weekly PDS Tickets Linked to PTRs",
            'font': {'size': 24, 'color': '#1a1a1a'},
            'y': 0.95,
            'x': 0.5,
            'xanchor': 'center',
            'yanchor': 'top'
        },
        xaxis_title={
            'text': "Week Starting",
            'font': {'size': 14, 'color': '#1a1a1a'}
        },
        yaxis_title={
            'text': "Number of New Linked Tickets",
            'font': {'size': 14, 'color': '#1a1a1a'}
        },
        barmode='stack',
        showlegend=False,
        height=800,
        plot_bgcolor='white',
        paper_bgcolor='white',
        hovermode='x unified',
        font={'family': "Arial, sans-serif"},
        margin={'t': 100, 'b': 100, 'l': 80, 'r': 80}
    )
    
    # Update axes
    fig.update_xaxes(
        tickangle=45,
        gridcolor='rgba(0, 0, 0, 0.2)',
        showgrid=True,
        zeroline=False,
        showline=True,
        linecolor='rgba(0, 0, 0, 0.5)',
        tickfont={'size': 12}
    )
    
    fig.update_yaxes(
        gridcolor='rgba(0, 0, 0, 0.2)',
        showgrid=True,
        zeroline=False,
        showline=True,
        linecolor='rgba(0, 0, 0, 0.5)',
        tickfont={'size': 12}
    )
    
    return fig

if __name__ == "__main__":
    try:
        ptr_path = "EXPORT/ptr.csv"
        pds_path = "EXPORT/pds.csv"
        fig = create_ptr_visualization(ptr_path, pds_path)
        fig.show()
    except Exception as e:
        print(f"Error creating visualization: {str(e)}")