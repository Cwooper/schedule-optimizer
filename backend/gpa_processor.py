# gpa_processor.py

import os
import pandas as pd
from Course import Course
from data_refresh import fetch_terms_list, filter_terms

current_directory = os.path.dirname(os.path.realpath(__file__))
data_directory = os.path.join(current_directory, 'data')
gpa_excel_file = os.path.join(data_directory, "Grade_Distribution_5_Year.xlsx")
gpa_pickle_file = os.path.join(data_directory, "grades.pkl")

columns_to_drop = ['TERM', 'CRN', 'TITLE', 'Students enrolled', 'Grade count']

def excel_to_pickle():
    if not os.path.exists(gpa_excel_file):
        print(f"Error: File not found: {gpa_excel_file}.")
        exit(1)
    df = pd.read_excel(gpa_excel_file)
    df = df.drop(columns_to_drop, axis=1)
    df.to_pickle(gpa_pickle_file)
    print(f"Created {gpa_pickle_file}.")

def main():
    if not os.path.exists(gpa_pickle_file):
        excel_to_pickle()
    gpa_df = pd.read_pickle(gpa_pickle_file)
    terms, year = filter_terms(fetch_terms_list())

    for term in terms:
        term_directory = os.path.join(data_directory, term)
        term_pkl = os.path.join(term_directory, term + '.pkl')
        term_df = pd.read_pickle(term_pkl)

        # courses = []
        # course_dicts = term_df.to_dict(orient='records')
        # if len(course_dicts) == 0:
        #     print("Error: empty dataframe")
        #     continue
        # 
        # for course_dict in course_dicts:
        #     courses.append(Course(**course_dicts))

        print(term_df)

if __name__ == "__main__":
    main()