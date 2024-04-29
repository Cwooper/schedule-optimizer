# gpa_processor.py

import os
import re
import numpy as np
import pandas as pd
from Course import Course
from data_refresh import fetch_terms_list, filter_terms

code_pattern = re.compile(r'(.*?)\s\d+')

current_directory = os.path.dirname(os.path.realpath(__file__))
data_directory = os.path.join(current_directory, 'data')
gpa_excel_file = os.path.join(data_directory, "Grade_Distribution_5_Year.xlsx")
gpa_pickle_file = os.path.join(data_directory, "grades.pkl")

columns_to_drop = ['TERM', 'CRN', 'TITLE', 'Students enrolled', 'Grade count']

def excel_to_pickle():
    if not os.path.exists(gpa_excel_file):
        print(f"Error: File not found: {gpa_excel_file}.")
        exit(1)
    print("Loading Excel...")
    df = pd.read_excel(gpa_excel_file)
    df = df.drop(columns_to_drop, axis=1)

    for index, row in df.iterrows():
        professor_name = row['PROFESSOR']
        if professor_name and isinstance(professor_name, str):
            professor_name_split = professor_name.split(' ')
            instructor_first = professor_name_split[:1]
            instructor_last = professor_name_split[-1:]
        else:
            instructor_first = None
            instructor_last = None
        
        df.at[index, 'INSTRUCTOR_FIRST'] = instructor_first[0] if instructor_first else None
        df.at[index, 'INSTRUCTOR_LAST'] = instructor_last[0] if instructor_last else None

    df = df.drop('PROFESSOR', axis=1)
    df.to_pickle(gpa_pickle_file)
    print(f"Created {gpa_pickle_file}.")

def calculate_gpa(row, gpa_df) -> int:
    code_match = re.match(code_pattern, row['subject'])
    code = None
    if code_match:
        code = code_match.group(1)

    instructor = row['instructor']

    first_name = None
    last_name = None

    if ', ' in instructor:
        names = instructor.split(',')
        last_names_str = names[0].strip()
        first_names_str = names[1].strip()

        last_names_list = last_names_str.split(' ')
        first_names_list = first_names_str.split(' ')

        first_name = first_names_list[:1]
        last_name = last_names_list[-1:]
    
    
    
    return 0.0

def main():
    if not os.path.exists(gpa_pickle_file):
        excel_to_pickle()
    gpa_df = pd.read_pickle(gpa_pickle_file)
    print(gpa_df)
    terms, year = filter_terms(fetch_terms_list())

    # Iterate over every term folder in data dir
    for term in terms:
        term_directory = os.path.join(data_directory, term)
        term_pkl = os.path.join(term_directory, term + '.pkl')
        term_df = pd.read_pickle(term_pkl)

        # Calculate the gpa for every row
        term_df['gpa'] = term_df.apply(calculate_gpa, axis=1, gpa_df=gpa_df)

        print(term_df)

if __name__ == "__main__":
    main()