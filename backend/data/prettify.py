from bs4 import BeautifulSoup

input = "ACCT.html"
output = input.replace(".html", "_pretty.html")

with open("ACCT.html", 'r', encoding='utf-8') as input_file:
    text = input_file.read()
    soup = BeautifulSoup(text, 'html.parser')

with open("ACCT_pretty.html", 'w', encoding='utf-8') as output_file:
    output_file.write(soup.prettify())

print(f"Succesffully prettified {input} to {output}")
