using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.IO;
using System.IO.Compression;
using System.Net;
using Newtonsoft.Json.Linq;
using System.Security.AccessControl;
using System.Security.Cryptography;
using System.Runtime.InteropServices;

namespace Installer
{
    class Installer
    {
        const int HWND_BROADCAST = 0xffff;
        const uint WM_SETTINGCHANGE = 0x001a;

        [DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
        static extern bool SendNotifyMessage(IntPtr hWnd, uint Msg,
            UIntPtr wParam, string lParam);

        static void Main(string[] args)
        {      
            Console.WriteLine("Getting latest release information from https://api.github.com/repos/cfschilham/kryer/releases/latest...");

            JObject release = getJson("https://api.github.com/repos/cfschilham/kryer/releases/latest");

            if (!Directory.Exists(@"C:\Program Files\Kryer"))
            {
                Directory.CreateDirectory(@"C:\Program Files\Kryer");
            }
            if(File.Exists(@"C:\Program Files\Kryer\kryer.exe"))
            {
                Console.Write("Kryer is already installed, reinstall/update [Y/n]: ");
                String q = Console.ReadLine();
                if(q.ToLower().Equals("n"))
                {
                    cleanup(true);
                    Console.ReadKey();
                    Environment.Exit(0);
                }
                Console.WriteLine(@"Removing C:\Program Files\Kryer\kryer.exe...");
                if(!getDirWritePerms(@"C:\Program Files\Kryer")) 
                {
                    Console.WriteLine("Insufficient permissions, run as admin");
                    cleanup(false);
                    Console.ReadKey();
                    Environment.Exit(1);
                }
                File.Delete(@"C:\Program Files\Kryer\kryer.exe");

                Console.WriteLine(@"Removed C:\Program Files\Kryer\kryer.exe...");

            }
            if (!Directory.Exists(@"C:\Program Files\Kryer\tmp"))
            {
                Directory.CreateDirectory(@"C:\Program Files\Kryer\tmp");
            }

            Console.WriteLine("Getting asset list from " + release["assets_url"] + "...");
            string name = string.Empty;
            foreach(JObject asset in release["assets"])
            {
                if(asset["name"].ToString().Contains("windows") && asset["name"].ToString().Contains(".zip"))
                {
                    Console.WriteLine("Downloading " + asset["browser_download_url"] + "...");
                    using (var client = new WebClient())
                    {
                        client.DownloadFile(asset["browser_download_url"].ToString(), @"C:\Program Files\Kryer\tmp\kryer.zip");
                    }
                    name = asset["name"].ToString().Replace(".zip", "");
                    break;
                }
            }
            

            foreach (JObject asset in release["assets"])
            {
                if (asset["name"].ToString().Equals(name + ".sha256"))
                {
                    Console.WriteLine("Getting SHA256 checksum from " + asset["browser_download_url"] + "...");
    
                    using (var client = new WebClient())
                    {
                        client.DownloadFile(asset["browser_download_url"].ToString(), @"C:\Program Files\Kryer\tmp\kryer.sha256");
                    }
                    break;
                }
            }
            string checksum = File.ReadAllText(@"C:\Program Files\Kryer\tmp\kryer.sha256").Split(" ".ToCharArray()[0])[0];
            Console.WriteLine("Verifying checksum " + checksum + "...");
            string filesum = string.Empty;
            using (SHA256 SHA256 = SHA256Managed.Create())
            {
                using (FileStream fileStream = File.OpenRead(@"C:\Program Files\Kryer\tmp\kryer.zip"))
                    filesum = BitConverter.ToString(SHA256.ComputeHash(fileStream)).Replace("-", "").ToLower();
            }
            if(checksum.Trim() != filesum.Trim())
            {
                Console.WriteLine("Integrity check failed, invalid checksum...");
                Console.WriteLine("Calculated checksum: " + filesum.Trim() + "...");
                Console.WriteLine("Provided checksum: " + checksum.Trim() + "...");
                cleanup(true);
                Console.ReadKey();
                Environment.Exit(1);
            }
            Console.WriteLine("Integrity check successful...");

            Console.WriteLine("Extracting " + name + ".zip...");

            ZipFile.ExtractToDirectory(@"C:\Program Files\Kryer\tmp\kryer.zip", @"C:\Program Files\Kryer\tmp\kryer");
            Console.WriteLine(@"Creating files in C:\Program Files\Kryer...");
            File.Move(@"C:\Program Files\Kryer\tmp\kryer\" + name + @"\kryer.exe", @"C:\Program Files\Kryer\kryer.exe");
            Console.WriteLine(@"Created C:\Program Files\Kryer\kryer.exe in C:\Program Files\Kryer...");



            string pathvar = Environment.GetEnvironmentVariable("PATH");
            if(!pathvar.Contains("Kryer"))
            {
                Console.WriteLine(@"Adding C:\Program Files\Kryer to PATH...");
                var value = pathvar + ";" + @"C:\Program Files\Kryer";
                var target = EnvironmentVariableTarget.Machine;
                Environment.SetEnvironmentVariable("PATH", value, target);
            } else
            {
                Console.WriteLine(@"C:\Program Files\Kryer already added to PATH...");
            }

            cleanup(true);
            Console.ReadKey();

        }

        public static void cleanup(Boolean verbose)
        {
            if(verbose) Console.WriteLine("Cleaning up...");

            if (Directory.Exists(@"C:\Program Files\Kryer\tmp")) DeleteDirectory(@"C:\Program Files\Kryer\tmp");
            if (verbose) Console.WriteLine("Done...");

        }

        public static bool getDirWritePerms(string path)
        {
            var writeAllow = false;
            var writeDeny = false;
            var accessControlList = Directory.GetAccessControl(path);
            if (accessControlList == null)
                return false;
            var accessRules = accessControlList.GetAccessRules(true, true,
                                        typeof(System.Security.Principal.SecurityIdentifier));
            if (accessRules == null)
                return false;

            foreach (FileSystemAccessRule rule in accessRules)
            {
                if ((FileSystemRights.Write & rule.FileSystemRights) != FileSystemRights.Write)
                    continue;

                if (rule.AccessControlType == AccessControlType.Allow)
                    writeAllow = true;
                else if (rule.AccessControlType == AccessControlType.Deny)
                    writeDeny = true;
            }

            return writeAllow && !writeDeny;
        }

        public static JObject getJson(string url)
        {
            string rawdata = string.Empty;
            HttpWebRequest request = (HttpWebRequest)WebRequest.Create(url);
            request.AutomaticDecompression = DecompressionMethods.GZip;
            request.UserAgent = ".NET Installer api";

            using (HttpWebResponse response = (HttpWebResponse)request.GetResponse())
            using (Stream stream = response.GetResponseStream())
            using (StreamReader reader = new StreamReader(stream))
            {
                rawdata = reader.ReadToEnd();
            }

            return JObject.Parse(rawdata);
        }

        public static void DeleteDirectory(string target_dir)
        {
            string[] files = Directory.GetFiles(target_dir);
            string[] dirs = Directory.GetDirectories(target_dir);

            foreach (string file in files)
            {
                File.SetAttributes(file, FileAttributes.Normal);
                File.Delete(file);
            }

            foreach (string dir in dirs)
            {
                DeleteDirectory(dir);
            }

            Directory.Delete(target_dir, false);
        }
    }
}
