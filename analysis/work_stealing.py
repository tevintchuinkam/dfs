import pandas as pd
import matplotlib.pyplot as plt

# Load the data
data = pd.read_csv("../results/workstealing.csv")

# Convert the data into a DataFrame
df = pd.DataFrame(data)

# Function to convert time strings to milliseconds
def convert_to_ms(time_str):
    if 'ms' in time_str:
        return float(time_str.replace('ms', ''))
    elif 's' in time_str:
        return float(time_str.replace('s', '')) * 1000
    else:
        raise ValueError("Unexpected time format")

# Apply the conversion function to the 'Time Taken' column
df['Time Taken'] = df['Time Taken'].apply(convert_to_ms)

# Compute the average time taken for each algorithm and folders per level
avg_time = df.groupby(['Algo', 'FoldersPerLevel'])['Time Taken'].mean().reset_index()

# Plot the data
plt.figure(figsize=(12, 6))

for algo in avg_time['Algo'].unique():
    algo_data = avg_time[avg_time['Algo'] == algo]
    plt.plot(algo_data['FoldersPerLevel'], algo_data['Time Taken'], marker='o', label=algo)

plt.xlabel('Folders Per Level')
plt.ylabel('Average Time Taken (ms)')
plt.title('Average Time Taken by Algorithm and Folders Per Level')
plt.legend()
plt.grid(True)
plt.show()
