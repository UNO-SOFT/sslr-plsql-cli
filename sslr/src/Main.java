package sslr;

import java.io.*;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.util.Collections;
import java.util.List;

import com.A.B.H.C;
import com.sonar.oracle.H;
import com.sonar.oracle.toolkit.A;
import com.sonar.sslr.api.AstNode;
import com.sonar.sslr.impl.ast.AstXmlPrinter;
import org.sonar.B.E;
import org.sonar.sslr.grammar.LexerlessGrammarBuilder;
import org.sonar.sslr.internal.toolkit.SourceCodeModel;
import org.sonar.sslr.parser.ParseRunner;
import org.sonar.sslr.parser.ParserAdapter;
import org.sonar.sslr.parser.ParsingResult;
import com.sonar.sslr.impl.Parser;
import org.sonar.sslr.toolkit.*;

public class Main {
    ParseRunner parseRunner;
    @C
    protected H B() {
        return com.sonar.oracle.H.B().A(
                StandardCharsets.UTF_8
        ).A(false).D();
    }

    public Main() {
        LexerlessGrammarBuilder builder = com.sonar.oracle.A.B();
        H h = this.B();
        /*
        this.parseRunner = new ParseRunner(
                com.sonar.oracle.I.A(
                        h,
                        (com.sonar.sslr.impl.Parser)(new org.sonar.sslr.parser.ParserAdapter(h.C(), builder.build())))
        );*/
    }

    public ParsingResult parse(char[] data ) {
        return this.parseRunner.parse(data);
    }

    public static void main(final String[] args) {
        ConfigurationModel cm = new com.sonar.oracle.toolkit.A();
        org.sonar.sslr.toolkit.Toolkit tk = new org.sonar.sslr.toolkit.Toolkit("SSLR :: PL/SQL :: Toolkit", cm);

        BufferedReader br = new BufferedReader(new InputStreamReader(System.in));
        try {
            StringWriter sw = new StringWriter(1024);
            br.transferTo(sw);
            String s = sw.toString();

            AstNode node = cm.getParser().parse(s);

            System.out.println("node=" + node);

            OutputStreamWriter osw = new OutputStreamWriter(System.out);
            AstXmlPrinter.print(node, osw);
            osw.flush();
        } catch(IOException ioe) {
            System.out.println(ioe);
        }
    }
}
