#include "parser/scanner.h"
#include "parser/parser.h"
#include <fstream>
#include <iostream>
#include <string>
#include <vector>

using namespace std;

void runFile(const string &path);
void runPrompt();
void run(const string &source);

int main(int argc, char *argv[]) {
  if (argc > 2) {
    cerr << "Usage: clox [script]" << endl;
    exit(64);
  } else if (argc == 2) {
    runFile(argv[1]);
  } else {
    runPrompt();
  }
  return 0;
}

void runFile(const string &path) {
  ifstream file(path);
  if (!file.is_open()) {
    cerr << "failed to open file: " << path << endl;
    exit(74);
  }

  string source((istreambuf_iterator<char>(file)), istreambuf_iterator<char>());
  run(source);
}

void runPrompt() {
  string line;
  cout << "> " << flush;

  while (getline(cin, line)) {
    run(line);
    cout << "> " << flush;
  }
}

void run(const string &source) {
  Scanner scanner(source);
  std::vector<Token> tokens;
  scanner.scan_tokens(tokens);
  if (scanner.has_error())
    return;
  for (auto &token : tokens) {
    std::cout << token.toString() << endl;
  }
  Parser parser(tokens);
  if (parser.has_error())
    return;
}
