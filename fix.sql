UPDATE account_codes 
SET code = 'S.' || code, 
    mak = replace(mak, code, 'S.' || code) 
WHERE code NOT LIKE 'S.%';
