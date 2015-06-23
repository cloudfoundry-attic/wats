// mmapleak.cpp : Defines the entry point for the console application.
//

#include "stdafx.h"
#include "windows.h"


int _tmain(int argc, _TCHAR* argv[])
{
	DWORD initialSize = 1024 * 1024 * 200;

	
	while (1){
		wchar_t name[1000];
		SYSTEMTIME st;
		GetSystemTime(&st);
		wsprintf(name, L"%02d-%02d-%02d", st.wMinute, st.wSecond, st.wMilliseconds);

		HANDLE hMapFile = CreateFileMapping(
			INVALID_HANDLE_VALUE,    // use paging file
			NULL,                    // default security
			PAGE_READWRITE,          // read/write access
			0,                       // maximum object size (high-order DWORD)
			initialSize,                // maximum object size (low-order DWORD)
			name);                  // name of mapping object


		if (hMapFile == NULL) {
			if (initialSize > 8) {
				initialSize = initialSize / 2;
			}
		}
		else {
			if (initialSize < 1024 * 1024 * 200) {
				initialSize = initialSize * 2;
			}
		}
	}

	return 0;
}

