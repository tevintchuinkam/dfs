import pandas as pd
import matplotlib.pyplot as plt


data = pd.read_csv("../results/workstealing.csv")

# Convert the data into a DataFrame
df = pd.DataFrame(data)

# Remove the 'ms' from the Time Taken and convert to float
df['Time Taken'] = df['Time Taken'].str.replace('ms', '').astype(float)

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