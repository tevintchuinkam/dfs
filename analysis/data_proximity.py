import pandas as pd
import matplotlib.pyplot as plt

# Creating the DataFrame
df = pd.read_csv("../results/data_proximity.csv")

# Converting Time Taken to seconds
def convert_time_to_seconds(time_str):
    if time_str.endswith('ms'):
        return float(time_str[:-2]) / 1000
    elif time_str.endswith('s'):
        return float(time_str[:-1])
    return float(time_str)

df["Time Taken (s)"] = df["Time Taken"].apply(convert_time_to_seconds)

# Compute the 95th percentile for each group
df['p95'] = df.groupby(['FileCount', 'DataProximity'])['Time Taken (s)'].transform(lambda x: x.quantile(0.95))

# Filter the values to include only those below the 95th percentile
df_filtered = df[df['Time Taken (s)'] <= df['p95']]

# Compute the average time taken for each algorithm and folders per level
df_mean = df_filtered.groupby(['FileCount', 'DataProximity'])['Time Taken (s)'].mean().reset_index()

# Plotting
plt.figure(figsize=(12, 6))

for proximity in df_mean["DataProximity"].unique():
    subset = df_mean[df_mean["DataProximity"] == proximity]
    plt.plot(subset["FileCount"], subset["Time Taken (s)"], marker='o', label=f'DataProximity={proximity}')

plt.xlabel('File Count')
plt.ylabel('Time Taken [seconds]')
plt.title('Time Taken vs File Count with and without Data Proximity')
plt.legend()
plt.grid(True)
plt.savefig("../results/data_proximity.png")
plt.show()
