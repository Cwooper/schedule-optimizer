# gpa_processor.py

import os
import re
import numpy as np
import pandas as pd

from data_refresh import fetch_terms_list, filter_terms

code_pattern = re.compile(r'(.*?)\s(\d+[A-Z]?)$')

current_directory = os.path.dirname(os.path.realpath(__file__))
data_directory = os.path.join(current_directory, 'data')
gpa_excel_file = os.path.join(data_directory, "Grade_Distribution_5_Year.xlsx")
gpa_pickle_file = os.path.join(data_directory, "grades.pkl")

columns_to_drop = ['TERM', 'CRN', 'TITLE', 'Students enrolled', 'Grade count', 'CNT_W'] 
grade_columns = ['CNT_A', 'CNT_AM', 'CNT_BP', 'CNT_B', 'CNT_BM', 'CNT_CP',
                 'CNT_C', 'CNT_CM', 'CNT_DP', 'CNT_D', 'CNT_DM', 'CNT_F']
gpa_values = np.array([4.0, 3.7, 3.3, 3.0, 2.7, 2.3, 2.0, 1.7, 1.3, 1.0, 0.7, 0.0])

mapping = {
    'AHE':'CFPA',
    'ASLC':'SPED',
    'ARAB':'LANG',
    'ASTR':'PHYS',
    'BNS':'PSY',
    'BUS':'ACCT',
    'CHIN':'LANG',
    'CLST':'LANG',
    'CD':'HHD',
    'CSEC':'CSE',
    'C2C':'HCS',
    'CISS':'CSCI',
    'DNC':'THTR',
    'DATA':'CSCI',
    'DIAD':'EDUC',
    'ECE':'ELED',
    'EDAD':'SPED',
    'ESJ':'SEC',
    'EECE':'ENGD',
    'ENGR':'ENGD',
    'EUS':'LANG',
    'FIN':'FMKT',
    'FREN':'LANG',
    'GERM':'LANG',
    'GLBL':'ESCI',
    'GREK':'LANG',
    'HLED':'HHD',
    'HRM':'MGMT',
    'HSP':'HCS',
    'HUMA':'GHR',
    'ID':'ENGD',
    'I T':'ECEM',
    'IEP':'LANG',
    'IBUS':'MGMT',
    'ITAL':'LANG',
    'JAPN':'LANG',
    'KIN':'HHD',
    'LAT':'LANG',
    'MIS':'DSCI',
    'MFGE':'ENGD',
    'MKTG':'FMKT',
    'MPAC':'ACCT',
    'M/CS':'MATH',
    'MLE':'ECEM',
    'NURS':'HCS',
    'OPS':'MBA',
    'PA':'HHD',
    'PE':'HHD',
    'PEH':'HHD',
    'PME':'ENGD',
    'PORT':'LANG',
    'RECR':'HHD',
    'RC':'HCS',
    'REL':'LBRL',
    'RUSS':'LANG',
    'SPAN':'LANG',
    'SUST':'UEPP',
    'TEOP':'ELIT',
    'TESL':'ELED'
}

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

def get_subject(code: str) -> str:
    subject_match = re.match(code_pattern, code)
    if subject_match:
        subject = subject_match.group(1)
        code = subject_match.group(2)

    # Replace any terms subjects with the associated in the gpa_df
    for key in mapping.keys():
        if key == subject:
            subject = subject.replace(key, mapping[key])
            break
    
    return subject + " " + code

# Returns instructor first_name, last_name if it is found
def get_instructor(instructor: str):
    first_name = None
    last_name = None
    if ', ' in instructor:
        names = instructor.split(',')
        last_names_str = names[0].strip()
        first_names_str = names[1].strip()

        last_names_list = last_names_str.split(' ')
        first_names_list = first_names_str.split(' ')

        first_name = first_names_list[0]
        last_name = last_names_list[-1]

    return first_name, last_name

def calculate_average_gpa(df):
    grade_counts = df[grade_columns].sum().tolist()

    if all(count == 0 for count in grade_counts):
        # If all grade_counts are zero return None
        return None
    average_gpa = np.dot(gpa_values, grade_counts) / np.sum(grade_counts)

    return round(average_gpa, 2)

def get_match_df(subject, first_name, last_name, gpa_df):
    # Find a direct Match
    match_df = gpa_df[(gpa_df['INSTRUCTOR_FIRST'].str.contains(first_name, regex=False)) &
                      (gpa_df['INSTRUCTOR_LAST'].str.contains(last_name, regex=False)) &
                      (gpa_df['CODE'] == subject)]
    
    if match_df.empty:
        return gpa_df[(gpa_df['INSTRUCTOR_LAST'].str.contains(last_name, regex=False)) &
                      (gpa_df['CODE'] == subject)]
    else:
        return match_df

# This will pass in a row from the terms_df, and calculate gpa based on gpa_df
def calculate_gpa(row, gpa_df):
    subject = get_subject(row['subject'])
    first_name, last_name = get_instructor(row['instructor'])
    match_df = get_match_df(subject, first_name, last_name, gpa_df)

    if match_df.empty:
        subject_df = gpa_df[(gpa_df['CODE'] == subject)]
        if subject_df.empty:
            instructor_df = gpa_df[(gpa_df['INSTRUCTOR_FIRST'].str.contains(first_name, regex=False)) & 
                                   (gpa_df['INSTRUCTOR_LAST'].str.contains(last_name, regex=False))]
            if instructor_df.empty:
                return None
            else:
                return calculate_average_gpa(instructor_df)
        else:
            return calculate_average_gpa(subject_df)
    else:
        return calculate_average_gpa(match_df)

def main():
    if not os.path.exists(gpa_pickle_file):
        excel_to_pickle()
    gpa_df = pd.read_pickle(gpa_pickle_file)
    gpa_df = gpa_df.drop('CNT_W', axis=1)
    gpa_df.dropna(subset=['CNT_A'], inplace=True)
    terms, _ = filter_terms(fetch_terms_list())

    # Iterate over every term folder in data dir
    for term in terms:
        print(f"Calculating {term}, this may take a minute...")
        term_directory = os.path.join(data_directory, term)
        term_pkl = os.path.join(term_directory, term + '.pkl')
        term_df = pd.read_pickle(term_pkl)

        # Calculate the gpa for every row
        term_df['gpa'] = term_df.apply(calculate_gpa, axis=1, gpa_df=gpa_df)
        term_df = term_df.replace(np.nan, None)
        term_df.to_pickle(term_pkl)
        print(f"Succesfully saved {term_pkl}!")

if __name__ == "__main__":
    main()