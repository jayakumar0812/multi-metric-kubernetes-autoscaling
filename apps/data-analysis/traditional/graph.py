#!/usr/bin/env python3
"""
SCALING COMPARISON GRAPH GENERATOR
Generates comparative graphs between traditional and intelligent scaling
"""

import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import numpy as np
from pathlib import Path
import sys

# Set style for professional graphs
plt.style.use('seaborn-v0_8')
sns.set_palette("husl")

def load_data(traditional_csv, intelligent_csv=None):
    """Load CSV data from traditional and optionally intelligent scaling tests"""
    
    # Load traditional data
    traditional_df = pd.read_csv(traditional_csv)
    traditional_df['System'] = 'Traditional (CPU-based)'
    
    # If intelligent data provided, load it
    if intelligent_csv and Path(intelligent_csv).exists():
        intelligent_df = pd.read_csv(intelligent_csv)
        intelligent_df['System'] = 'Intelligent (Complexity-based)'
        combined_df = pd.concat([traditional_df, intelligent_df], ignore_index=True)
    else:
        # Use your actual research output data
        print("📊 No intelligent data file provided, using actual research output data...")
        intelligent_data = {
            'Load_Level': ['low', 'medium', 'high', 'very_high'],
            'Complexity_Score': [126, 368, 795, 878],
            'Pod_Count_Before': [2, 2, 2, 10],
            'Pod_Count_After': [2, 2, 10, 10],
            'Wait_Time': [210, 225, 270, 285],
            'Scaling_Result': ['NO_SCALING', 'NO_SCALING', 'SCALED_UP', 'SCALED_UP'],
            'System': ['Intelligent (Complexity-based)'] * 4
        }
        intelligent_df = pd.DataFrame(intelligent_data)
        combined_df = pd.concat([traditional_df, intelligent_df], ignore_index=True)
    
    return combined_df

def create_scaling_comparison_graph(df, output_dir):
    """Create comprehensive scaling comparison graphs"""
    
    # Create figure with subplots
    fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(16, 12))
    fig.suptitle('Traditional vs Intelligent Auto-Scaling Comparison', fontsize=16, fontweight='bold')
    
    # Graph 1: Pod Scaling by Load Level
    load_order = ['low', 'medium', 'high', 'very_high']
    df['Load_Level'] = pd.Categorical(df['Load_Level'], categories=load_order, ordered=True)
    
    scaling_data = df.pivot_table(
        values='Pod_Count_After', 
        index='Load_Level', 
        columns='System', 
        aggfunc='mean'
    ).reindex(load_order)
    
    scaling_data.plot(kind='bar', ax=ax1, width=0.7)
    ax1.set_title('Pod Count After Scaling by Load Level', fontweight='bold')
    ax1.set_xlabel('Load Level')
    ax1.set_ylabel('Number of Pods')
    ax1.legend(title='Scaling System')
    ax1.grid(True, alpha=0.3)
    
    # Graph 2: Complexity Score vs Scaling Response
    for system in df['System'].unique():
        system_data = df[df['System'] == system]
        colors = ['red' if result == 'NO_SCALING' else 'green' for result in system_data['Scaling_Result']]
        ax2.scatter(system_data['Complexity_Score'], system_data['Pod_Count_After'], 
                   label=system, alpha=0.7, s=100, c=colors)
    
    ax2.set_title('Complexity Score vs Pod Count (Red=No Scaling, Green=Scaled)', fontweight='bold')
    ax2.set_xlabel('Complexity Score')
    ax2.set_ylabel('Pods After Scaling')
    ax2.legend()
    ax2.grid(True, alpha=0.3)
    
    # Graph 3: Scaling Efficiency (Pods Added vs Load Level)
    df['Pods_Added'] = df['Pod_Count_After'] - df['Pod_Count_Before']
    
    efficiency_data = df.pivot_table(
        values='Pods_Added', 
        index='Load_Level', 
        columns='System', 
        aggfunc='mean'
    ).reindex(load_order)
    
    efficiency_data.plot(kind='bar', ax=ax3, width=0.7)
    ax3.set_title('Scaling Efficiency: Pods Added by Load Level', fontweight='bold')
    ax3.set_xlabel('Load Level')
    ax3.set_ylabel('Pods Added')
    ax3.legend(title='Scaling System')
    ax3.grid(True, alpha=0.3)
    
    # Graph 4: Response Time Analysis
    response_data = df.pivot_table(
        values='Wait_Time', 
        index='Load_Level', 
        columns='System', 
        aggfunc='mean'
    ).reindex(load_order)
    
    response_data.plot(kind='line', ax=ax4, marker='o', linewidth=2, markersize=8)
    ax4.set_title('Scaling Response Time by Load Level', fontweight='bold')
    ax4.set_xlabel('Load Level')
    ax4.set_ylabel('Wait Time (seconds)')
    ax4.legend(title='Scaling System')
    ax4.grid(True, alpha=0.3)
    
    plt.tight_layout()
    
    # Save the graph
    output_path = Path(output_dir) / 'scaling_comparison.png'
    plt.savefig(output_path, dpi=300, bbox_inches='tight')
    print(f"📊 Comparison graph saved: {output_path}")
    
    return fig

def create_detailed_analysis_graph(df, output_dir):
    """Create detailed analysis graphs"""
    
    fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(16, 12))
    fig.suptitle('Detailed Scaling Analysis', fontsize=16, fontweight='bold')
    
    # Graph 1: Scaling Success Rate
    scaling_success = df.groupby(['System', 'Load_Level'])['Scaling_Result'].apply(
        lambda x: (x == 'SCALED_UP').sum() / len(x) * 100
    ).unstack(fill_value=0)
    
    scaling_success.plot(kind='bar', ax=ax1, width=0.8)
    ax1.set_title('Scaling Success Rate by Load Level (%)', fontweight='bold')
    ax1.set_xlabel('Scaling System')
    ax1.set_ylabel('Success Rate (%)')
    ax1.legend(title='Load Level', bbox_to_anchor=(1.05, 1), loc='upper left')
    ax1.grid(True, alpha=0.3)
    
    # Graph 2: Complexity Score Distribution
    for system in df['System'].unique():
        system_data = df[df['System'] == system]
        ax2.hist(system_data['Complexity_Score'], alpha=0.7, label=system, bins=10)
    
    ax2.set_title('Complexity Score Distribution', fontweight='bold')
    ax2.set_xlabel('Complexity Score')
    ax2.set_ylabel('Frequency')
    ax2.legend()
    ax2.grid(True, alpha=0.3)
    
    # Graph 3: Load Level Impact
    load_impact = df.groupby(['Load_Level', 'System']).agg({
        'Pod_Count_After': 'mean',
        'Complexity_Score': 'mean'
    }).reset_index()
    
    traditional_data = load_impact[load_impact['System'].str.contains('Traditional')]
    intelligent_data = load_impact[load_impact['System'].str.contains('Intelligent')]
    
    x = np.arange(len(traditional_data))
    width = 0.35
    
    ax3.bar(x - width/2, traditional_data['Pod_Count_After'], width, 
            label='Traditional', alpha=0.8)
    ax3.bar(x + width/2, intelligent_data['Pod_Count_After'], width, 
            label='Intelligent', alpha=0.8)
    
    ax3.set_title('Average Pod Count by Load Level', fontweight='bold')
    ax3.set_xlabel('Load Level')
    ax3.set_ylabel('Average Pod Count')
    ax3.set_xticks(x)
    ax3.set_xticklabels(traditional_data['Load_Level'])
    ax3.legend()
    ax3.grid(True, alpha=0.3)
    
    # Graph 4: Scaling Responsiveness Heatmap
    responsiveness = df.pivot_table(
        values=['Wait_Time', 'Pods_Added'], 
        index='Load_Level', 
        columns='System'
    )
    
    # Normalize the data for heatmap
    wait_time_norm = responsiveness['Wait_Time'].div(responsiveness['Wait_Time'].max(axis=1), axis=0)
    
    sns.heatmap(wait_time_norm, annot=True, cmap='RdYlGn_r', ax=ax4, fmt='.2f')
    ax4.set_title('Scaling Response Time (Normalized)', fontweight='bold')
    ax4.set_xlabel('Scaling System')
    ax4.set_ylabel('Load Level')
    
    plt.tight_layout()
    
    # Save the graph
    output_path = Path(output_dir) / 'detailed_analysis.png'
    plt.savefig(output_path, dpi=300, bbox_inches='tight')
    print(f"📊 Detailed analysis graph saved: {output_path}")
    
    return fig

def create_research_summary_graph(df, output_dir):
    """Create a research paper quality summary graph"""
    
    fig, ax = plt.subplots(1, 1, figsize=(12, 8))
    
    # Calculate key metrics
    summary_stats = df.groupby('System').agg({
        'Pod_Count_After': ['mean', 'std'],
        'Wait_Time': ['mean', 'std'],
        'Complexity_Score': 'mean'
    }).round(2)
    
    # Create summary visualization
    systems = df['System'].unique()
    load_levels = ['low', 'medium', 'high', 'very_high']
    
    x = np.arange(len(load_levels))
    width = 0.35
    
    for i, system in enumerate(systems):
        system_data = df[df['System'] == system]
        pods = [system_data[system_data['Load_Level'] == level]['Pod_Count_After'].mean() 
                for level in load_levels]
        
        ax.bar(x + i*width, pods, width, label=system, alpha=0.8)
    
    ax.set_title('GraphQL Auto-Scaling Comparison: Research Results', 
                fontsize=14, fontweight='bold', pad=20)
    ax.set_xlabel('Query Complexity Level', fontsize=12)
    ax.set_ylabel('Average Pod Count', fontsize=12)
    ax.set_xticks(x + width/2)
    ax.set_xticklabels([level.title() for level in load_levels])
    ax.legend(fontsize=11)
    ax.grid(True, alpha=0.3, axis='y')
    
    # Add performance annotations
    for i, system in enumerate(systems):
        system_data = df[df['System'] == system]
        scaling_events = (system_data['Scaling_Result'] == 'SCALED_UP').sum()
        total_tests = len(system_data)
        success_rate = (scaling_events / total_tests) * 100
        
        ax.text(0.02 + i*0.5, 0.95 - i*0.05, 
                f'{system}: {success_rate:.0f}% scaling success rate',
                transform=ax.transAxes, fontsize=10,
                bbox=dict(boxstyle="round,pad=0.3", facecolor='lightblue', alpha=0.7))
    
    plt.tight_layout()
    
    # Save the graph
    output_path = Path(output_dir) / 'research_summary.png'
    plt.savefig(output_path, dpi=300, bbox_inches='tight')
    print(f"📊 Research summary graph saved: {output_path}")
    
    return fig

def generate_performance_report(df, output_dir):
    """Generate a text report of performance metrics"""
    
    report_path = Path(output_dir) / 'performance_report.txt'
    
    with open(report_path, 'w') as f:
        f.write("GRAPHQL AUTO-SCALING PERFORMANCE REPORT\n")
        f.write("="*50 + "\n\n")
        
        for system in df['System'].unique():
            system_data = df[df['System'] == system]
            f.write(f"🔧 {system}\n")
            f.write("-" * 30 + "\n")
            
            # Basic stats
            f.write(f"Average pods after scaling: {system_data['Pod_Count_After'].mean():.1f}\n")
            f.write(f"Average response time: {system_data['Wait_Time'].mean():.1f}s\n")
            f.write(f"Average complexity handled: {system_data['Complexity_Score'].mean():.1f}\n")
            
            # Scaling efficiency
            scaling_events = (system_data['Scaling_Result'] == 'SCALED_UP').sum()
            total_tests = len(system_data)
            f.write(f"Scaling success rate: {(scaling_events/total_tests)*100:.1f}%\n")
            
            # Pod utilization
            total_pods_added = (system_data['Pod_Count_After'] - system_data['Pod_Count_Before']).sum()
            f.write(f"Total pods added: {total_pods_added}\n")
            f.write(f"Average pods per scaling event: {total_pods_added/max(scaling_events,1):.1f}\n")
            
            f.write("\n")
        
        # Comparison
        f.write("📊 COMPARISON ANALYSIS\n")
        f.write("-" * 30 + "\n")
        
        if len(df['System'].unique()) == 2:
            traditional = df[df['System'].str.contains('Traditional')]
            intelligent = df[df['System'].str.contains('Intelligent')]
            
            trad_success = (traditional['Scaling_Result'] == 'SCALED_UP').sum() / len(traditional) * 100
            intel_success = (intelligent['Scaling_Result'] == 'SCALED_UP').sum() / len(intelligent) * 100
            
            f.write(f"Intelligent system scaling success rate: {intel_success:.1f}%\n")
            f.write(f"Traditional system scaling success rate: {trad_success:.1f}%\n")
            f.write(f"Improvement: {intel_success - trad_success:.1f} percentage points\n")
            
            avg_response_intel = intelligent['Wait_Time'].mean()
            avg_response_trad = traditional['Wait_Time'].mean()
            f.write(f"Response time improvement: {avg_response_trad - avg_response_intel:.1f}s faster\n")
    
    print(f"📋 Performance report saved: {report_path}")

def main():
    """Main function to generate all graphs"""
    
    if len(sys.argv) < 2:
        print("Usage: python scaling_graph_generator.py <traditional_csv> [intelligent_csv]")
        print("Example: python scaling_graph_generator.py localhost_results.csv intelligent_results.csv")
        sys.exit(1)
    
    traditional_csv = sys.argv[1]
    intelligent_csv = sys.argv[2] if len(sys.argv) > 2 else None
    
    # Create output directory
    output_dir = Path("scaling_graphs")
    output_dir.mkdir(exist_ok=True)
    
    print("🚀 GENERATING SCALING COMPARISON GRAPHS")
    print("=" * 40)
    print(f"📊 Traditional data: {traditional_csv}")
    if intelligent_csv:
        print(f"📊 Intelligent data: {intelligent_csv}")
    else:
        print("📊 Intelligent data: Using sample data for comparison")
    print(f"📁 Output directory: {output_dir}")
    print()
    
    # Load data
    try:
        df = load_data(traditional_csv, intelligent_csv)
        print(f"✅ Data loaded successfully: {len(df)} records")
        print(f"📊 Systems: {df['System'].unique()}")
        print()
    except Exception as e:
        print(f"❌ Error loading data: {e}")
        sys.exit(1)
    
    # Generate graphs
    try:
        print("📈 Generating comparison graphs...")
        create_scaling_comparison_graph(df, output_dir)
        
        print("📈 Generating detailed analysis...")
        create_detailed_analysis_graph(df, output_dir)
        
        print("📈 Generating research summary...")
        create_research_summary_graph(df, output_dir)
        
        print("📋 Generating performance report...")
        generate_performance_report(df, output_dir)
        
        print()
        print("🎉 ALL GRAPHS GENERATED SUCCESSFULLY!")
        print("=" * 40)
        print(f"📁 Check the '{output_dir}' directory for:")
        print("   📊 scaling_comparison.png - Main comparison graphs")
        print("   📊 detailed_analysis.png - Detailed analysis")
        print("   📊 research_summary.png - Research paper quality graph")
        print("   📋 performance_report.txt - Performance metrics report")
        
    except Exception as e:
        print(f"❌ Error generating graphs: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()