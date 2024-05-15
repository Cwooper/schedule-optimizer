import pandas as pd

df = pd.read_pickle('data/202440/202440.pkl')

print(df[(df['subject'].str.contains('FAIR'))])