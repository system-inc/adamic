#ifndef ADAMIC_LIBRARY_NUMBER_FORMAT_H
#define ADAMIC_LIBRARY_NUMBER_FORMAT_H
adamic_string *adamic_number_checked_fixed(double value, double digits);
adamic_string *adamic_number_checked_exponential(double value, double digits);
adamic_string *adamic_number_checked_precision(double value, double digits);
adamic_string *adamic_number_checked_radix(double value, double radix);
#endif
