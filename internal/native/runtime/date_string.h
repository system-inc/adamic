#ifndef ADAMIC_DATE_STRING_H
#define ADAMIC_DATE_STRING_H
// Borrow the existing Date internal slot; the formatted result is owned.
adamic_string *adamic_fs_file_date_string(const adamic_object *date);
#endif
