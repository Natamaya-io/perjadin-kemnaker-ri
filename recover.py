import json
import re
import sys

def main():
    log_file = r"C:\Users\ramvi\.gemini\tmp\perjadin-kemnaker-ri-1\tool-outputs\session-765b22e7-4974-4300-99cb-2a5274f33075\replace_replace_1778257716396_0_edxyt9.txt"
    with open(log_file, "r", encoding="utf-8") as f:
        data = json.load(f)
    
    output = data.get("output", "")
    marker = "Here is the updated code:\n"
    idx = output.find(marker)
    if idx != -1:
        code = output[idx + len(marker):]
        with open(r"D:\Projects\web-project\perjadin-kemnaker-ri\backend\internal\domain\record\repository.go", "w", encoding="utf-8") as f:
            f.write(code)
        print("Recovered successfully!")
    else:
        print("Marker not found.")

if __name__ == "__main__":
    main()