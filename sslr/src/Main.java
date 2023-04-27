package sslr;

import java.io.*;
import java.net.InetSocketAddress;
import java.util.Iterator;

import com.sonar.sslr.api.AstNode;
import com.sonar.sslr.impl.ast.AstXmlPrinter;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;
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
        //(new BufferedReader(r)).transferTo(sw);
        char[] buffer = new char[4096];
        int len;
        while((len = r.read(buffer)) >= 0) {
            sw.write(buffer, 0, len);
        }

        sw.flush();
        String s = sw.toString();
        return this.parse(s);
    }
    public AstNode parse(InputStream in) throws IOException {
        return this.parse(new InputStreamReader(in));
    }

    public void printNode(OutputStream os, AstNode node) throws IOException {
        OutputStreamWriter osw = new OutputStreamWriter(os);
        //AstXmlPrinter.print(node, osw);
        Main.printNodes(osw, 0, node);
        osw.flush();
        osw.close();
        os.flush();
    }

    private static void printNodes(Writer w, int indent, AstNode node) throws IOException {
        if (indent != 0) {
            w.append("\n");
        }

        Main.printNodeIndent(w, indent);
        if (node.hasChildren()) {
            w.append("<");
            Main.printNodeValue(w, node);
            w.append(">");
            Main.printNodeChildren(w, indent, node);
            Main.printNodeNL(w, indent);
            w.append("</").append(node.getName()).append(">");
        } else {
            w.append("<");
            Main.printNodeValue(w, node);
            w.append("/>");
        }
    }

    private static void printNodeValue(Writer w, AstNode node) throws IOException {
        w.append(node.getName());
        if (node.getTokenValue() != null) {
            w.append(" tokenValue=\"").
                    append(org.apache.commons.text.StringEscapeUtils.escapeXml11(node.getTokenValue())).
                    append("\"");
        }

        if (node.hasToken()) {
            w.append(" tokenLine=\"").
                    append(String.valueOf(node.getTokenLine())).
                    append("\" tokenColumn=\"").
                    append(String.valueOf(node.getToken().getColumn())).
                    append("\"");
        }
    }

    private static void printNodeChildren(Writer w, int indent, AstNode node) throws IOException {
        for (AstNode child : node.getChildren()) {
            Main.printNodes(w, indent + 1, child);
        }
    }

    private static void printNodeNL(Writer w, int indent) throws IOException {
        w.append("\n");
        printNodeIndent(w, indent);
    }

    private static void printNodeIndent(Writer w, int indent) throws IOException {
        int i = 0;
        for(i = 0; i < indent; i++) {
            w.append("  ");
        }
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
