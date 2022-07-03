package codequerypairsimporter

import (
	"context"
	"testing"

	"github.com/hexops/autogold"
)

func TestSOQuestionToCodeQueryPair(t *testing.T) {
	pythonCodeHTML := `<code>
# Comment
def a():
	"""Sum"""
	# Comment
	return 1 + 1
</code>`

	pythonCodeWithMultipleTagsHTML := `<code>aReallyLongFunctionCall(1, 2)</code>
<p>Paragraph</p>
<code>
a = 1
b = 2
c = a - b
</code>
`

	goCodeWithMultipleTagsHTML := `<p>Paragraph</p>
<code>
func f() {
	// We do this
	1+1
	...
	// After a while we do it again
	1+1
}
</code>
`

	php := `<pre><code><?php
/**
	* Handler for persistent data files
	*
	* @author Ryan McCue <cubegames@gmail.com>
	* @package Lilina
	* @version 1.0
	* @license http://opensource.org/licenses/gpl-license.php GNU Public License
	*/

/**
	* Handler for persistent data files
	*
	* @package Lilina
	*/
class DataHandler {
	/**
		* Directory to store data.
		*
		* @since 1.0
		*
		* @var string
		*/
	protected $directory;

	/**
		* Constructor, duh.
		*
		* @since 1.0
		* @uses $directory Holds the data directory, which the constructor sets.
		*
		* @param string $directory 
		*/
	public function __construct($directory = null) {
		if ($directory === null)
			$directory = get_data_dir();

		if (substr($directory, -1) != '/')
			$directory .= '/';

		$this->directory = (string) $directory;
	}

	/**
		* Prepares filename and content for saving
		*
		* @since 1.0
		* @uses $directory
		* @uses put()
		*
		* @param string $filename Filename to save to
		* @param string $content Content to save to cache
		*/
	public function save($filename, $content) {
		$file = $this->directory . $filename;

		if(!$this->put($file, $content)) {
			trigger_error(get_class($this) . " error: Couldn't write to $file", E_USER_WARNING);
			return false;
		}

		return true;
	}
}

?>
</code></pre>

<p>It stores each entry as a separate file, which we found is efficient enough for use (no unneeded data is loaded and it's faster to save).</p>`

	phpWithoutTags := `<p>With <a href="http://php.net/manual/en/class.datetime.php" rel="noreferrer"><code>DateTime</code> API</a>:</p>

<pre><code>$dateTime = new DateTime('2008-09-22'); 
echo $dateTime->format('U'); 

// or 

$date = new DateTime('2008-09-22');
echo $date->getTimestamp();
</code></pre>

<p>The same with the procedural API:</p>

<pre><code>$date = date_create('2008-09-22');
echo date_format($date, 'U');

// or

$date = date_create('2008-09-22');
echo date_timestamp_get($date);
</code></pre>

<p>If the above fails because you are using a <a href="http://www.php.net/manual/de/datetime.formats.php" rel="noreferrer">unsupported format</a>, you can use</p>

<pre><code>$date = DateTime::createFromFormat('!d-m-Y', '22-09-2008');
echo $dateTime->format('U'); 

// or

$date = date_parse_from_format('!d-m-Y', '22-09-2008');
echo date_format($date, 'U');
</code></pre>

<p>Note that if you do not set the <code>!</code>, the time portion will be set to current time, which is different from the first four which will use midnight when you omit the time.</p>

<p>Yet another alternative is to use the <a href="http://php.net/manual/en/class.intldateformatter.php" rel="noreferrer"><code>IntlDateFormatter</code></a> API:</p>

<pre><code>$formatter = new IntlDateFormatter(
	'en_US',
	IntlDateFormatter::FULL,
	IntlDateFormatter::FULL,
	'GMT',
	IntlDateFormatter::GREGORIAN,
	'dd-MM-yyyy'
);
echo $formatter->parse('22-09-2008');
</code></pre>

<p>Unless you are working with localized date strings, the easier choice is likely DateTime.</p>`

	tests := []struct {
		name string
		q    SOQuestionWithAnswers
	}{
		{
			name: "Question with single code answer",
			q:    SOQuestionWithAnswers{Title: "Title 1", Tags: "<unk><sec><python>", Answers: []*string{&pythonCodeHTML}},
		},
		{
			name: "Question with wrong tag for code answer",
			q:    SOQuestionWithAnswers{Title: "Title 1", Tags: "<java><sec><python>", Answers: []*string{&pythonCodeHTML}},
		},
		{
			name: "Question with multiple code answers",
			q:    SOQuestionWithAnswers{Title: "Title 1", Tags: "<python><unk>", Answers: []*string{&pythonCodeHTML, &pythonCodeWithMultipleTagsHTML}},
		},
		{
			name: "Code answers with dots in the middle",
			q:    SOQuestionWithAnswers{Title: "Title 1", Tags: "<go>", Answers: []*string{&goCodeWithMultipleTagsHTML}},
		},
		{
			name: "PHP weird",
			q:    SOQuestionWithAnswers{Title: "Title 1", Tags: "<php>", Answers: []*string{&php}},
		},
		{
			name: "PHP without tags",
			q:    SOQuestionWithAnswers{Title: "Title 1", Tags: "<php>", Answers: []*string{&phpWithoutTags}},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cqp, err := questionToCodeQueryPair(ctx, nil, &tt.q, false)
			if err != nil {
				t.Fatal(err)
			}
			// Ignore in the autogold snapshot.
			cqp.SOQuestionID = nil
			autogold.Equal(t, cqp)
		})
	}
}
