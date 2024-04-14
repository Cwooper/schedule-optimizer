import pandas as pd

df = pd.read_pickle("202420.pkl")
df.to_csv("202420.csv", index=False)

