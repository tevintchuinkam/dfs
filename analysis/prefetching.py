import pandas as pd
import matplotlib.pyplot as plt

# Load the CSV file
df = pd.read_csv("../results/prefetching.csv")

# Convert UseCache to boolean type if it is not already
df['UseCache'] = df['UseCache'].astype('bool')

# Convert Time Taken from nanoseconds to milliseconds
df['Time Taken'] = df['Time Taken'] / 1_000_000

# Separate DataFrames based on UseCache
df_true = df[df['UseCache'] == True]
df_false = df[df['UseCache'] == False]

def process_data(df, percentile):
    # Create a call index within each iteration
    df['Call Index'] = df.groupby('Iteration').cumcount()
    
    # Pivot the table to have iterations as columns and call indices as rows
    pivot_df = df.pivot(index='Call Index', columns='Iteration', values='Time Taken')
    
    # Compute the specified percentile across the iterations (axis=1)
    p95_times = pivot_df.quantile(percentile, axis=1)
    
    # Convert it to DataFrame for ease of saving to CSV
    p95_df = p95_times.reset_index()
    p95_df.columns = ['Call Index', 'P95 Time Taken']

    return p95_df

# Specify the percentile
percentile = 0.90

# Process data for true and false UseCache values
p95_df_true = process_data(df_true, percentile)
p95_df_false = process_data(df_false, percentile)

# Save the 95th percentile times to new CSV files
p95_df_true.to_csv("../results/p95_true_milliseconds.csv", index=False)
p95_df_false.to_csv("../results/p95_false_milliseconds.csv", index=False)

# Calculate the average of the 95th percentile times
average_p95_true = p95_df_true['P95 Time Taken'].mean()
average_p95_false = p95_df_false['P95 Time Taken'].mean()

# Print the averages
print(f"Average P95 Time Taken (UseCache=True): {average_p95_true} ms")
print(f"Average P95 Time Taken (UseCache=False): {average_p95_false} ms")

# Plot the results
plt.figure(figsize=(10, 6))

plt.plot(p95_df_true['Call Index'], p95_df_true['P95 Time Taken'], marker='o', linestyle='-', label='UseCache=True')
plt.plot(p95_df_false['Call Index'], p95_df_false['P95 Time Taken'], marker='x', linestyle='-', label='UseCache=False')

plt.xlabel('Call Index')
plt.ylabel('Time Taken [milliseconds]')
plt.title('Time Taken for Each Call Over Iterations')
plt.legend()
plt.grid(True)
plt.savefig("../results/p95_plot_overlay_milliseconds.png")
plt.show()
