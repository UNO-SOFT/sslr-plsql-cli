package sslr;

import java.io.*;
import java.nio.charset.Charset;

import com.A.B.H.C;
import com.sonar.oracle.H;
import org.sonar.sslr.parser.ParseRunner;
import org.sonar.sslr.parser.ParsingResult;
import com.sonar.sslr.impl.Parser;
import org.sonar.sslr.toolkit.ConfigurationProperty;
import org.sonar.sslr.toolkit.Validators;

public class Main {
    ParseRunner parseRunner;
    @C
    protected H B() {
        return com.sonar.oracle.H.B().A(
                Charset.forName("UTF-8")
        ).A(true).D();
    }

    public Main() {
        this.parseRunner = new ParseRunner(
                com.sonar.oracle.I.A(this.B()).getRootRule()
        );
    }

    public ParsingResult parse(char[] data ) {
        return this.parseRunner.parse(data);
    }

    public static void main(final String[] args) {
        BufferedReader br = new BufferedReader(new InputStreamReader(System.in));
        char []data = null;
        try {
            CharArrayWriter caw = new CharArrayWriter(1024);
            br.transferTo(caw);
            data = caw.toCharArray();
        } catch(IOException ioe) {
            System.out.println(ioe);
        }
        Main m = new Main();

        ParsingResult res = m.parse(data);
        System.out.println("res=" + res);
    }
}
