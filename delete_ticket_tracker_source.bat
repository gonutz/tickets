for /F %%f in ('git ls-tree -r master --name-only') do del "%%f"
rmdir /S /Q .git