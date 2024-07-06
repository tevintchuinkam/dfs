import pandas as pd
import matplotlib.pyplot as plt


data = pd.read_csv("../results/results_workstealing.csv")

# Convert the data into a DataFrame
df = pd.DataFrame(data)

# Convert time to a consistent format (seconds)
def convert_to_seconds(time_str):
    if 's' in time_str and 'ms' not in time_str:
        return float(time_str.replace('s', ''))
    elif 'ms' in time_str:
        return float(time_str.replace('ms', '')) / 1000.0

df['Time Taken (s)'] = df['Time Taken'].apply(convert_to_seconds)

# Calculate average time taken by each algorithm
average_time = df.groupby('Algo')['Time Taken (s)'].mean().reset_index()

# Print the average times
print("Average Time Taken by Each Algorithm:")
print(average_time)

# Plot the comparison
plt.figure(figsize=(10, 6))
for algo in df['Algo'].unique():
    algo_data = df[df['Algo'] == algo]
    plt.plot(algo_data['Iteration'], algo_data['Time Taken (s)'], marker='o', label=algo)

plt.title('Performance Comparison of Algorithms')
plt.xlabel('Iteration')
plt.ylabel('Time Taken (seconds)')
plt.legend()
plt.grid(True)
plt.show()