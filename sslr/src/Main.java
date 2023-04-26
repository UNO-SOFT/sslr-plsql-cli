package sslr;

import java.io.*;
import java.net.InetSocketAddress;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.util.Collections;
import java.util.List;

import com.A.B.H.C;
import com.sonar.oracle.H;
import com.sonar.oracle.toolkit.A;
import com.sonar.sslr.api.AstNode;
import com.sonar.sslr.impl.ast.AstXmlPrinter;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;
import org.sonar.B.E;
import org.sonar.sslr.grammar.LexerlessGrammarBuilder;
import org.sonar.sslr.internal.toolkit.SourceCodeModel;
import org.sonar.sslr.parser.ParseRunner;
import org.sonar.sslr.parser.ParserAdapter;
import org.sonar.sslr.parser.ParsingResult;
import com.sonar.sslr.impl.Parser;
import org.sonar.sslr.toolkit.*;

public class Main implements HttpHandler {
    private ConfigurationModel cm = new com.sonar.oracle.toolkit.A();

    public Main() {
        //org.sonar.sslr.toolkit.Toolkit tk = new org.sonar.sslr.toolkit.Toolkit("SSLR :: PL/SQL :: Toolkit", cm);
    }

    public AstNode parse(String s) {
        return this.cm.getParser().parse(s);
    }
    public AstNode parse(Reader r) throws IOException {
        StringWriter sw = new StringWriter(1024);
        (new BufferedReader(r)).transferTo(sw);
        sw.flush();
        String s = sw.toString();
        return this.parse(s);
    }
    public AstNode parse(InputStream in) throws IOException {
        return this.parse(new InputStreamReader(in));
    }

    public void printNode(OutputStream os, AstNode node) throws IOException {
        OutputStreamWriter osw = new OutputStreamWriter(os);
        AstXmlPrinter.print(node, osw);
        osw.flush();
        osw.close();
        os.flush();
    }

    public void handle(HttpExchange t) throws IOException {
        System.out.println(t.getRequestMethod()+" "+t.getRequestURI()+" from "+t.getRemoteAddress());
        AstNode node = this.parse(t.getRequestBody());
        System.out.println("node: "+node);
        t.getResponseHeaders().set("Content-Type", "application/xml");
        t.sendResponseHeaders(200, 0);
        OutputStream os = t.getResponseBody();
        this.printNode(os, node);
        os.flush();
        os.close();
    }

    public static void main(final String[] args) {
        Main m = new Main();
        try {
            System.out.println(args.length);
            if(args.length > 1 && args[0].equals("-port")) {
                int port = Integer.parseInt(args[1]);
                System.out.println("Listening on localhost:"+port);
                HttpServer server = HttpServer.create(new InetSocketAddress(port), 5);
                server.createContext("/", m);
                server.setExecutor(null); // creates a default executor
                server.start();
                return;
            }
            AstNode node = m.parse(System.in);
            m.printNode(System.out, node);
        } catch(IOException ioe) {
            System.out.println("ERROR:" + ioe);
        }
    }
}
