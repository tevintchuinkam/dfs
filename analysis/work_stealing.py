import pandas as pd
import matplotlib.pyplot as plt

# Load the data
data = pd.read_csv("../results/workstealing.csv")

# Convert the data into a DataFrame
df = pd.DataFrame(data)

# Function to convert time strings to milliseconds
def convert_to_seconds(time_str):
    if 'ms' in time_str:
        return float(time_str.replace('ms', '')) / 1000
    elif 's' in time_str:
        return float(time_str.replace('s', '')) 
    else:
        raise ValueError("Unexpected time format")

# Apply the conversion function to the 'Time Taken' column
df['Time Taken'] = df['Time Taken'].apply(convert_to_seconds)

# Compute the 95th percentile for each algorithm and added latency
df_p95 = df.groupby(['Algo', 'AddedLatency'])['Time Taken'].quantile(0.90).reset_index()
df_p95.rename(columns={'Time Taken': 'P95 Time Taken'}, inplace=True)

# Compute the average of the 95th percentile times for each algorithm and added latency
avg_p95_time = df_p95.groupby(['Algo', 'AddedLatency'])['P95 Time Taken'].mean().reset_index()

# Plot the data
plt.figure(figsize=(12, 6))

for algo in avg_p95_time['Algo'].unique():
    algo_data = avg_p95_time[avg_p95_time['Algo'] == algo]
    plt.plot(algo_data['AddedLatency'], algo_data['P95 Time Taken'], marker='o', label=algo)

plt.xlabel('Added Latency per Request [milliseconds]')
plt.ylabel('Time Taken [seconds]')
plt.title('Time Taken by Algorithm for Directory Traversal')
plt.legend()
plt.grid(True)
plt.savefig("../results/work_stealing.png")
plt.show()
