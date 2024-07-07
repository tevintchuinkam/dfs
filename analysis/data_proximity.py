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

# Plotting
plt.figure(figsize=(10, 6))

for proximity in df["DataProximity"].unique():
    subset = df[df["DataProximity"] == proximity]
    plt.plot(subset["FileSizeMB"], subset["Time Taken (s)"], marker='o', label=f'DataProximity={proximity}')

plt.xlabel('FileSize (MB)')
plt.ylabel('Time Taken (s)')
plt.title('Time Taken vs FileSize with Data Proximity')
plt.legend()
plt.grid(True)
plt.savefig("../results/data_proximity.png")
plt.show()