#!/bin/bash

# Script to create a sample PDF for testing
# Requires: pdflatex or any text-to-PDF tool

SAMPLE_FILE="sample.pdf"

echo "📄 Creating sample PDF for testing..."

# Create a temporary tex file
cat > /tmp/sample.tex << 'EOF'
\documentclass{article}
\usepackage[utf8]{inputenc}
\title{Sample PDF Document}
\author{Test User}
\date{\today}

\begin{document}

\maketitle

\section{Introduction}
This is a sample PDF document created for testing the Multi-Tenant PDF Ingestion Service.

\section{Main Content}
This document contains multiple paragraphs of text to test the PDF extraction and AI summarization capabilities.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

\section{Conclusion}
This concludes our sample PDF document. The service should be able to extract this text and generate a meaningful summary.

\end{document}
EOF

# Try to compile with pdflatex if available
if command -v pdflatex &> /dev/null; then
    pdflatex -interaction=nonstopmode -output-directory=/tmp /tmp/sample.tex > /dev/null 2>&1
    mv /tmp/sample.pdf ./$SAMPLE_FILE
    echo "✅ Sample PDF created: $SAMPLE_FILE"
else
    echo "⚠️  pdflatex not found. Please create a sample PDF manually and name it '$SAMPLE_FILE'"
    echo "   Or you can use any existing PDF file for testing."
fi

# Cleanup
rm -f /tmp/sample.tex /tmp/sample.aux /tmp/sample.log

