import pandas as pd
import matplotlib.pyplot as plt

# Load the data
data = pd.read_csv("../results/workstealing_micro.csv")

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

# Compute the average time taken for each algorithm and folders per level
avg_time = df.groupby(['Algo', 'LatencyMS'])['Time Taken'].median().reset_index()

# Plot the data
plt.figure(figsize=(12, 6))

for algo in avg_time['Algo'].unique():
    algo_data = avg_time[avg_time['Algo'] == algo]
    plt.plot(algo_data['LatencyMS'], algo_data['Time Taken'], marker='o', label=algo)

plt.xlabel('Latency per Request (milliseconds)')
plt.ylabel('Average Time Taken (seconds)')
plt.title('Average Time Taken by Algorithm and Folders Per Level')
plt.legend()
plt.grid(True)
plt.savefig("../results/work_stealing.png")
plt.show()
