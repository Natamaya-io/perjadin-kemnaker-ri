import zipfile
import re
import shutil
import os

docx_path = 'backend/templates/Berkas Luar Kota - Laporan.docx'
backup_path = 'backend/templates/Berkas Luar Kota - Laporan_backup.docx'
shutil.copy(docx_path, backup_path)

with zipfile.ZipFile(backup_path, 'r') as zin:
    with zipfile.ZipFile(docx_path, 'w') as zout:
        for item in zin.infolist():
            content = zin.read(item.filename)
            if item.filename == 'word/document.xml':
                xml = content.decode('utf-8')
                
                # Find row 7
                m = re.search(r'<w:tr[^>]*>(?:(?!</w:tr>).)*?tujuh(?:(?!</w:tr>).)*?</w:tr>', xml)
                if m:
                    row7 = m.group(0)
                    row8 = row7.replace('tujuh', 'delapan')
                    row9 = row7.replace('tujuh', 'sembilan')
                    row10 = row7.replace('tujuh', 'sepuluh')
                    
                    new_xml = xml.replace(row7, row7 + row8 + row9 + row10)
                    content = new_xml.encode('utf-8')
                else:
                    print("Row tujuh not found!")
            zout.writestr(item, content)

print("Done fixing DOCX.")
