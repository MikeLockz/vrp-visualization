import pandas as pd
import plotly.graph_objects as go
from plotly.subplots import make_subplots
from collections import Counter
from datetime import datetime, timedelta

def count_issues_by_labels(df, start_date=None, end_date=None):
    df_filtered = df
    if start_date and end_date:
        df_filtered = df[(df['Created'] >= start_date) & (df['Created'] <= end_date)]
    
    label_columns = [col for col in df.columns if col == 'Labels']
    all_labels = []
    
    for _, row in df_filtered.iterrows():
        for label_col in label_columns:
            label = row[label_col]
            if pd.notna(label) and str(label).strip():
                all_labels.append(str(label).strip())
    
    return Counter(all_labels), df_filtered

def create_interactive_visualization(csv_file, top_n=15):
    df = pd.read_csv(csv_file, low_memory=False)
    df['Created'] = pd.to_datetime(df['Created'], format='mixed')
    
    max_date = df['Created'].max()
    
    date_ranges = {
        '1 Day': (max_date - timedelta(days=1), max_date),
        '1 Week': (max_date - timedelta(weeks=1), max_date),
        '1 Month': (max_date - timedelta(days=30), max_date),
        '3 Months': (max_date - timedelta(days=90), max_date),
        '6 Months': (max_date - timedelta(days=180), max_date),
        '1 Year': (max_date - timedelta(days=365), max_date),
        'All Time': (df['Created'].min(), max_date)
    }
    
    fig = make_subplots(
        rows=2, cols=2,
        specs=[[{"type": "indicator"}, {"type": "indicator"}],
               [{"type": "bar"}, {"type": "pie"}]],
        subplot_titles=('Total Tickets', 'Average Tickets per Day',
                       'Top Issue Labels', 'Distribution of Top Issue Labels'),
        row_heights=[0.2, 0.8]
    )
    
    def update_plots(start_date, end_date):
        label_counts, df_filtered = count_issues_by_labels(df, start_date, end_date)
        
        plot_df = pd.DataFrame.from_dict(label_counts, orient='index', columns=['count'])
        plot_df = plot_df.sort_values('count', ascending=True).tail(top_n)
        
        total = plot_df['count'].sum()
        plot_df['percentage'] = plot_df['count'] / total * 100
        
        total_tickets = len(df_filtered)
        days_in_period = (end_date - start_date).days + 1
        avg_tickets_per_day = total_tickets / days_in_period
        
        return plot_df, total_tickets, avg_tickets_per_day
    
    initial_df, initial_total, initial_avg = update_plots(
        date_ranges['All Time'][0], 
        date_ranges['All Time'][1]
    )
    
    fig.add_trace(
        go.Indicator(
            mode="number",
            value=initial_total,
            title={"text": "Total Tickets"},
            number={'font': {'size': 50}, 'valueformat': ',d'},
        ),
        row=1, col=1
    )
    
    fig.add_trace(
        go.Indicator(
            mode="number",
            value=initial_avg,
            title={"text": "Average Tickets per Day"},
            number={'font': {'size': 50}, 'valueformat': ',.1f'},
        ),
        row=1, col=2
    )
    
    fig.add_trace(
        go.Bar(
            x=initial_df['count'],
            y=initial_df.index,
            orientation='h',
            text=initial_df['count'],
            textposition='auto',
            name='Count'
        ),
        row=2, col=1
    )
    
    fig.add_trace(
        go.Pie(
            values=initial_df['count'],
            labels=initial_df.index,
            textinfo='percent+label',
            hole=0.3
        ),
        row=2, col=2
    )
    
    fig.update_layout(
        title_text="Issue Label Analysis",
        showlegend=False,
        height=1000,
        width=1400,
        updatemenus=[
            dict(
                type="dropdown",
                x=0.05,
                y=1.15,
                buttons=[
                    dict(
                        args=[{
                            'x': [update_plots(start, end)[0]['count']],
                            'y': [update_plots(start, end)[0].index],
                            'values': [update_plots(start, end)[0]['count']],
                            'labels': [update_plots(start, end)[0].index],
                            'value': [[update_plots(start, end)[1]], 
                                    [update_plots(start, end)[2]]]
                        }],
                        label=range_name,
                        method="update"
                    )
                    for range_name, (start, end) in date_ranges.items()
                ],
            )
        ]
    )
    
    fig.update_xaxes(title_text="Number of Issues", row=2, col=1)
    fig.update_yaxes(title_text="Labels", row=2, col=1)
    
    fig.write_html('interactive_label_distribution.html')
    print("\nInteractive plot saved as 'interactive_label_distribution.html'")
    
    print("\nAll-Time Label Statistics:")
    print("=" * 50)
    print(f"Total number of tickets: {initial_total:,}")
    print(f"Average tickets per day: {initial_avg:.1f}")
    print(f"Number of unique labels: {len(initial_df)}")
    print(f"\nTop {top_n} labels account for "
          f"{initial_df['count'].sum() / sum(initial_df['count'].values()) * 100:.1f}% of all labels")

if __name__ == "__main__":
    csv_file = 'EXPORT/pds.csv'
    
    try:
        create_interactive_visualization(csv_file)
    except Exception as e:
        print(f"An error occurred: {str(e)}")